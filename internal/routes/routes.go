package routes

import (
	"net/http"
	"time"

	oauth "github.com/haikali3/gymbara-backend/internal/auth"
	"github.com/haikali3/gymbara-backend/internal/controllers"
	"github.com/haikali3/gymbara-backend/internal/middleware"
	"github.com/haikali3/gymbara-backend/internal/payment"
	"github.com/haikali3/gymbara-backend/internal/payment/webhook"
)

// RegisterRoutes sets up all the HTTP routes for the application.
// It applies middleware for security, rate limiting, CORS, and authentication
// to ensure proper access control and request handling. The routes are grouped
// into the following categories:
//
// 1. Workout Routes:
//    - Handles workout-related endpoints such as fetching workout sections,
//      exercises, and their details. These routes require a subscription.
//
// 2. User Routes:
//    - Includes endpoints for submitting user exercise details, fetching user
//      progress, and retrieving user information.
//
// 3. OAuth Routes:
//    - Provides endpoints for handling OAuth login and callback functionality
//      using Google authentication.
//
// 4. Payment Routes:
//    - Manages payment-related endpoints such as creating subscriptions,
//      verifying checkout sessions, canceling subscriptions, and retrieving
//      subscription details.
//
// 5. Webhook Routes:
//    - Handles Stripe webhook events, such as checkout session completion.
//
// Middleware is applied to ensure proper security and functionality for each
// route group.

func RegisterRoutes() {
	const maxRequests = 10
	const duration = time.Second

	secureHandler := func(handler http.HandlerFunc) http.Handler {
		rateLimitedHandler := middleware.RateLimit(maxRequests, duration)
		corsHandler := middleware.CORS
		authHandler := middleware.AuthMiddleware

		return corsHandler(rateLimitedHandler(authHandler(handler)))
	}

	// Workout routes
	http.Handle("/workout-sections", secureHandler(middleware.RequireSubscription(controllers.GetWorkoutSections)))
	http.Handle("/workout-sections/list", secureHandler(middleware.RequireSubscription(controllers.GetExercisesList)))
	http.Handle("/workout-sections/details", secureHandler(middleware.RequireSubscription(controllers.GetExerciseDetails)))
	http.Handle("/workout-sections/exercises", secureHandler(middleware.RequireSubscription(controllers.GetWorkoutSectionsWithExercises)))

	// User submit exercise details
	http.Handle("/workout-sections/user-exercise-details", secureHandler(controllers.SubmitUserExerciseDetails))
	// Fetch user submitted exercise detail
	http.Handle("/user/progress", secureHandler(controllers.GetUserProgress))
	// Fetch user details
	http.Handle("/api/user-info", secureHandler(controllers.GetUserInfoHandler))

	// OAuth routes
	http.Handle("/oauth/login", middleware.RateLimit(maxRequests, duration)(http.HandlerFunc(oauth.GoogleLoginHandler)))
	http.Handle("/oauth/callback", middleware.RateLimit(maxRequests, duration)(http.HandlerFunc(oauth.GoogleCallbackHandler)))

	// Payment
	http.Handle("/payment/checkout", middleware.CORS(http.HandlerFunc(payment.CreateSubscription)))
	http.Handle("/payment/verify-session", middleware.CORS(http.HandlerFunc(payment.VerifyCheckoutSession)))
	http.Handle("/payment/cancel-subscription", secureHandler(payment.CancelSubscription))
	http.Handle("/payment/get-subscription", secureHandler(payment.GetSubscription))
	http.Handle("/payment/renew-subscription", secureHandler(payment.RenewSubscription))

	// Webhook
	http.Handle("/webhook/stripe", http.HandlerFunc(webhook.CheckoutSessionCompleted))
}
