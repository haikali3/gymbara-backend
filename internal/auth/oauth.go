// REFER https://www.reddit.com/r/golang/comments/10sqggb/how_to_implement_oauth_in_go/
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/internal/models"
	"github.com/haikali3/gymbara-backend/internal/utils"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuth scopes
const (
	ScopeUserInfoProfile = "https://www.googleapis.com/auth/userinfo.profile"
	ScopeUserInfoEmail   = "https://www.googleapis.com/auth/userinfo.email"
)

// GoogleOauthConfig stores OAuth2 configuration
var GoogleOauthConfig *oauth2.Config

// initializes the OAuth configuration using env variables
func InitializeOAuthConfig() error {
	backendBaseURL := os.Getenv("BACKEND_BASE_URL")
	if backendBaseURL == "" {
		utils.Logger.Fatal("BACKEND_BASE_URL is not set in the environment variables")
	}

	GoogleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  fmt.Sprintf("%s/oauth/callback", backendBaseURL),
		Scopes:       []string{ScopeUserInfoProfile, ScopeUserInfoEmail},
		Endpoint:     google.Endpoint,
	}

	utils.Logger.Info("OAuth configuration initialized",
		zap.String("redirect_url", GoogleOauthConfig.RedirectURL),
	)
	return nil
}

// creates a state token to prevent CSRF attacks -> stores it in a cookie
func GenerateStateOAuthCookie(w http.ResponseWriter) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b) //ignore error

	oauthStateString := base64.URLEncoding.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    oauthStateString,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		Secure:   true,
		HttpOnly: true, // ? change if https?
	})

	utils.Logger.Debug("Generated OAuth state cookie", zap.String("oauth_state", oauthStateString))
	return oauthStateString
}

// redirects the user to Google’s OAuth login
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	if GoogleOauthConfig == nil {
		if err := InitializeOAuthConfig(); err != nil {
			utils.Logger.Fatal("Failed to initialize OAuth config", zap.Error(err))
		}
	}

	oauthStateString := GenerateStateOAuthCookie(w)
	authURL := GoogleOauthConfig.AuthCodeURL(oauthStateString)
	utils.Logger.Info("Redirecting to Google OAuth URL", zap.String("auth_url", authURL))
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// handles Google OAuth callback and stores user info
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	//prevent csrf
	if !validateOAuthState(r) {
		stateCookie, err := r.Cookie("oauthstate")
		if err != nil {
			utils.Logger.Error("Failed to retrieve OAuth state cookie", zap.Error(err))
		} else {
			utils.Logger.Error("Invalid OAuth state",
				zap.String("state", r.FormValue("state")),
				zap.String("cookie_value", stateCookie.Value),
			)
		}
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	//exchg code for token from google's oauth2 server
	token, err := GoogleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		utils.Logger.Error("Error exchanging code for token", zap.Error(err))
		http.Error(w, "Could not get token", http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("OAuth state validated", zap.String("state", r.FormValue("state")))

	//get user info from google's api
	utils.Logger.Info("Fetching user info for token", zap.String("access_token", token.AccessToken)) // Add this log
	userInfo, err := fetchUserInfo(context.Background(), token)
	if err != nil {
		utils.Logger.Error("Error fetching user info", zap.Error(err))
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}

	err = database.StoreUserWithToken(userInfo, token.AccessToken)
	if err != nil {
		utils.Logger.Error("Error storing user in DB", zap.Error(err))
		http.Error(w, "Failed to store user info", http.StatusInternalServerError)
		return
	}

	// set session cookie with access token
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token.AccessToken,
		Expires:  time.Now().Add(time.Until(token.Expiry)), // This is more readable
		HttpOnly: true,
		Secure:   true, // Set to true if using HTTPS
		Path:     "/",
	})
	utils.Logger.Info("Session cookie set", zap.String("user_email", userInfo.Email))

	http.Redirect(w, r, getFrontendURL(), http.StatusSeeOther)
}

//TODO: create get access token
//TODO: create refresh token

// retrieves the user's info from Google
func fetchUserInfo(ctx context.Context, token *oauth2.Token) (models.GoogleUser, error) {
	client := GoogleOauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return models.GoogleUser{}, fmt.Errorf("error fetching user info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.GoogleUser{}, fmt.Errorf("error validating token: status code %v", resp.StatusCode)
	}

	var userInfo models.GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return models.GoogleUser{}, fmt.Errorf("error decoding user info: %v", err)
	}

	utils.Logger.Debug("Fetched user info from Google", zap.String("user_email", userInfo.Email))
	return userInfo, nil
}

func ValidateToken(accessToken string) (int, error) {
	token := &oauth2.Token{AccessToken: accessToken}
	userInfo, err := fetchUserInfo(context.Background(), token)
	if err != nil {
		utils.Logger.Error("Token validation failed", zap.Error(err))
		return 0, err
	}

	//fetch user ID from db using email
	var userID int
	err = database.DB.QueryRow("SELECT id FROM users WHERE email = $1", userInfo.Email).Scan(&userID)

	if err != nil {
		utils.Logger.Error("User not found",
			zap.String("user_email", userInfo.Email),
			zap.Error(err))
		return 0, fmt.Errorf("error fetching user ID: %v", err)
	}

	utils.Logger.Info("Token validated successfully",
		zap.String("user_email", userInfo.Email),
		zap.Int("user_id", userID))

	return userID, nil
}

// Helper function to validate OAuth state for CSRF protection
func validateOAuthState(r *http.Request) bool {
	stateCookie, err := r.Cookie("oauthstate")
	if err != nil || r.FormValue("state") != stateCookie.Value {
		utils.Logger.Error("Invalid OAuth state or cookie mismatch", zap.Error(err))
		return false
	}
	utils.Logger.Debug("Valid OAuth state", zap.String("state", r.FormValue("state")))
	return true
}

// Helper function to get frontend URL from environment variables
func getFrontendURL() string {
	return os.Getenv("FRONTEND_URL")
}
