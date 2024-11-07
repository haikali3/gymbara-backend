// REFER https://www.reddit.com/r/golang/comments/10sqggb/how_to_implement_oauth_in_go/
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/haikali3/gymbara-backend/database"
	"github.com/haikali3/gymbara-backend/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuth scopes
// ! Consent from user is needed
const (
	ScopeUserInfoProfile = "https://www.googleapis.com/auth/userinfo.profile"
	ScopeUserInfoEmail   = "https://www.googleapis.com/auth/userinfo.email"
)

// GoogleOauthConfig stores OAuth2 configuration
var GoogleOauthConfig *oauth2.Config

// initializes the OAuth configuration using env variables
func InitializeOAuthConfig() {
	backendBaseURL := os.Getenv("BACKEND_BASE_URL")
	if backendBaseURL == "" {
		log.Fatal("BACKEND_BASE_URL is not set in the environment variables")
	}

	GoogleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  fmt.Sprintf("%s/oauth/callback", backendBaseURL),
		Scopes:       []string{ScopeUserInfoProfile, ScopeUserInfoEmail},
		Endpoint:     google.Endpoint,
	}

	fmt.Printf("OAuth configuration initialized with Redirect URL: %s\n", GoogleOauthConfig.RedirectURL)
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

	return oauthStateString
}

// redirects the user to Google’s OAuth login
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	if GoogleOauthConfig == nil {
		InitializeOAuthConfig()
	}

	oauthStateString := GenerateStateOAuthCookie(w)
	authURL := GoogleOauthConfig.AuthCodeURL(oauthStateString)
	fmt.Printf("Redirecting to Google OAuth URL: %s\n", authURL)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// handles Google OAuth callback and stores user info
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	//prevent csrf
	if !validateOAuthState(r) {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	//exchg code for token from google's oauth2 server
	token, err := GoogleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		log.Printf("Error exchanging code for token: %v\n", err)
		http.Error(w, "Could not get token", http.StatusInternalServerError)
		return
	}

	//get user info from google's api
	userInfo, err := fetchUserInfo(context.Background(), token)
	if err != nil {
		log.Printf("Error fetching user info: %v\n", err)
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}

	err = database.StoreUserWithToken(userInfo, token.AccessToken)
	if err != nil {
		log.Printf("Error storing user in DB: %v\n", err)
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

	http.Redirect(w, r, getFrontendURL(), http.StatusSeeOther)
}

// retrieves the user's info from Google
func fetchUserInfo(ctx context.Context, token *oauth2.Token) (models.GoogleUser, error) {
	client := GoogleOauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return models.GoogleUser{}, fmt.Errorf("error fetching user info: %v", err)
	}
	defer resp.Body.Close()

	var userInfo models.GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return models.GoogleUser{}, fmt.Errorf("error decoding user info: %v", err)
	}

	return userInfo, nil
}

// clears the session cookie and redirects the user to the frontend
func GoogleLogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	})

	http.Redirect(w, r, getFrontendURL(), http.StatusSeeOther)
}

func GetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	authCookie, err := r.Cookie("access_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken := authCookie.Value
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken) //Need access token to call this API

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch use details from Google", http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	var userInfo models.GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Error decoding user info", http.StatusInternalServerError)
		return
	}

	// Return user data as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// Helper function to validate OAuth state for CSRF protection
func validateOAuthState(r *http.Request) bool {
	stateCookie, err := r.Cookie("oauthstate")
	if err != nil || r.FormValue("state") != stateCookie.Value {
		log.Println("Invalid OAuth state or cookie mismatch:", err)
		return false
	}
	return true
}

// Helper function to get frontend URL from environment variables
func getFrontendURL() string {
	return os.Getenv("FRONTEND_URL")
}
