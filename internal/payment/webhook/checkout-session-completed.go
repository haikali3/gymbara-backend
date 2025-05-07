// internal/payment/webhook/checkout-session-completed.go
package webhook

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/webhook"
	"go.uber.org/zap"
)

func CheckoutSessionCompleted(w http.ResponseWriter, r *http.Request) {
	// 1) Panic guard
	defer func() {
		if rec := recover(); rec != nil {
			utils.Logger.Error("panic in webhook handler", zap.Any("panic", rec))
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
	}()

	// 2) Limit payload size
	const MaxBodyBytes = 64 << 10
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	// 3) Read & verify
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		utils.Logger.Error("Error reading request body", zap.Error(err))
		http.Error(w, "Error reading request body", http.StatusServiceUnavailable)
		return
	}
	event, err := webhook.ConstructEvent(
		payload,
		r.Header.Get("Stripe-Signature"),
		os.Getenv("STRIPE_WEBHOOK_SECRET"),
	)
	if err != nil {
		utils.Logger.Error("Invalid webhook signature", zap.Error(err))
		http.Error(w, "Invalid webhook signature", http.StatusBadRequest)
		return
	}

	// 4) Acknowledge immediately
	w.WriteHeader(http.StatusOK)

	// 5) Process in background
	go func() {
		if event.Type != "checkout.session.completed" && event.Type != "invoice.payment_succeeded" {
			return
		}

		var (
			rawEmail string
			subID    string
			ts       int64
		)

		switch event.Type {
		case "checkout.session.completed":
			var sess stripe.CheckoutSession
			if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
				utils.Logger.Error("parse session failed", zap.Error(err))
				return
			}
			ts = sess.Created

			// ─────── Fallback for Subscription ID ───────
			if sess.Subscription != nil {
				subID = sess.Subscription.ID
			}
			if subID == "" && sess.Subscription != nil {
				subID = sess.Subscription.ID
			}

			// ─────── Fallback for Email ───────
			rawEmail = sess.CustomerEmail
			if rawEmail == "" && sess.CustomerDetails != nil {
				rawEmail = sess.CustomerDetails.Email
			}
			if rawEmail == "" && sess.Metadata != nil {
				rawEmail = sess.Metadata["email"]
			}

		case "invoice.payment_succeeded":
			var inv stripe.Invoice
			if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
				utils.Logger.Error("parse invoice failed", zap.Error(err))
				return
			}
			ts = inv.Created
			subID = inv.Subscription.ID
			rawEmail = inv.CustomerEmail
		}

		// 🌟 Debug log: show what we actually got
		utils.Logger.Info("Webhook data",
			zap.String("event_type", string(event.Type)),
			zap.String("subID", subID),
			zap.String("email", rawEmail),
		)

		// guard against missing data
		if subID == "" || rawEmail == "" {
			utils.Logger.Error("missing data; skipping upsert",
				zap.String("subID", subID),
				zap.String("email", rawEmail),
			)
			return
		}

		handleByEmail(subID, rawEmail, ts)
	}()
}

// handleByEmail upserts User (by email) and Subscription rows.
func handleByEmail(subID, email string, ts int64) {
	if email == "" {
		utils.Logger.Error("Missing customer email; skipping", zap.String("subID", subID))
		return
	}

	// fetch real billing period end
	sub, err := subscription.Get(subID, nil)
	if err != nil {
		utils.Logger.Error("Stripe subscription lookup failed", zap.Error(err))
		return
	}
	expiration := time.Unix(sub.CurrentPeriodEnd, 0)

	// begin transaction
	tx, err := database.DB.Begin()
	if err != nil {
		utils.Logger.Error("DB tx begin failed", zap.Error(err))
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			utils.Logger.Error("Failed to rollback transaction", zap.Error(err))
		}
	}()

	// upsert user (same as before)…
	var userID int
	if err := tx.QueryRow("SELECT id FROM Users WHERE email=$1", email).Scan(&userID); err != nil {
		// insert new user
		if err := tx.QueryRow(
			"INSERT INTO Users (email, is_premium) VALUES ($1, TRUE) RETURNING id",
			email,
		).Scan(&userID); err != nil {
			utils.Logger.Error("Failed to create user", zap.Error(err))
			return
		}
	} else {
		// mark existing premium
		if _, err := tx.Exec("UPDATE Users SET is_premium=TRUE WHERE id=$1", userID); err != nil {
			utils.Logger.Error("Failed to mark user premium", zap.Error(err))
			return
		}
	}

	// ───── plain INSERT/UPDATE for Subscriptions ─────
	var existingSubID string
	err = tx.QueryRow(
		`SELECT stripe_subscription_id
				FROM Subscriptions
			WHERE user_id = $1`,
		userID,
	).Scan(&existingSubID)

	if err == sql.ErrNoRows {
		// no row yet → INSERT
		if _, err := tx.Exec(
			`INSERT INTO Subscriptions
					(user_id, paid_date, expiration_date, stripe_subscription_id)
				VALUES ($1, $2, $3, $4)`,
			userID, time.Unix(ts, 0), expiration, subID,
		); err != nil {
			utils.Logger.Error("Failed to insert subscription", zap.Error(err))
			return
		}
	} else if err == nil {
		// row exists → UPDATE
		if _, err := tx.Exec(
			`UPDATE Subscriptions
							SET paid_date = $1,
									expiration_date = $2,
									stripe_subscription_id = $3
						WHERE user_id = $4`,
			time.Unix(ts, 0), expiration, subID, userID,
		); err != nil {
			utils.Logger.Error("Failed to update subscription", zap.Error(err))
			return
		}
	} else {
		// unexpected DB error
		utils.Logger.Error("Failed to query subscription", zap.Error(err))
		return
	}

	// commit
	if err := tx.Commit(); err != nil {
		utils.Logger.Error("DB commit failed", zap.Error(err))
		return
	}

	utils.Logger.Info("Upserted subscription by email",
		zap.String("email", email),
		zap.String("subID", subID),
		zap.Time("expires", expiration),
	)
}
