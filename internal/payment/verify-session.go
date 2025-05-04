// internal/payment/verify-session.go

package payment

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/haikali3/gymbara-backend/pkg/utils"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"go.uber.org/zap"
)

func VerifyCheckoutSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Missing session_id", http.StatusBadRequest)
		return
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	s, err := session.Get(sessionID, nil)
	if err != nil {
		utils.Logger.Error("Failed to retrieve session", zap.Error(err))
		http.Error(w, "Invalid session_id", http.StatusBadRequest)
		return
	}

	// Build frontend-friendly response
	resp := map[string]interface{}{
		"orderId": s.Subscription.ID,
		"date":    time.Now().Format("2006-01-02"),
		"items": []map[string]string{
			{
				"name":  "Gymbara Pro Membership",
				"price": "10.00MYR",
			},
		},
		"total": "10.00MYR",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		utils.Logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
