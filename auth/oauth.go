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

	"github.com/haikali3/gymbara-backend/database"
	"github.com/haikali3/gymbara-backend/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// config for google OAuth2
var GoogleOauthConfig = &oauth2.Config{
	RedirectURL: "http://localhost:8080/oauth/callback",
	Scopes:      []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:    google.Endpoint,
}

// security to prevent CSRF attacks
func GenerateStateOAuthCookie(w http.ResponseWriter) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b) //ignore error
	oauthStateString := base64.URLEncoding.EncodeToString(b)

	fmt.Printf("Generated OAuth state: %s", oauthStateString)

	http.SetCookie(w, &http.Cookie{
		Name:    "oauthstate",
		Value:   oauthStateString,
		Expires: time.Now().Add(24 * time.Hour),
		Path:    "/", //enable cookie avalable across all paths
	})

	return oauthStateString
}

// 1. Redirects user to Google login
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	// load from .env
	GoogleOauthConfig.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleOauthConfig.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")

	fmt.Println("Google Client ID (Handler): ", GoogleOauthConfig.ClientID)
	fmt.Println("Google Client Secret (Handler): ", GoogleOauthConfig.ClientSecret)

	// generate state, set it in cookie, add it to AuthCodeURL
	oauthStateString := GenerateStateOAuthCookie(w)
	url := GoogleOauthConfig.AuthCodeURL(oauthStateString)
	fmt.Println("Redirecting to Google OAuth URL: ", url)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// 2. Handles Google OAuth callback and processes user info
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	//check csrf with state param
	stateCookie, err := r.Cookie("oauthstate")
	if err != nil || r.FormValue("state") != stateCookie.Value {
		fmt.Println("Invalid OAuth state or cookie mismatch: \n", err)
		http.Error(w, "Invalid OAuth state cookie %v\n", http.StatusBadRequest)
		return
	}
	fmt.Println("OAuth State received: ", stateCookie.Value)

	//exchange auth code for access token
	token, err := GoogleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		fmt.Println("Error exchanging code for token: \n", err)
		http.Error(w, "Could not get token %v\n", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Access Token: %s\n", token.AccessToken)

	// get user info from google
	client := GoogleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		fmt.Println("Error fetching user info: \n", err)
		http.Error(w, "Failed to fetch user info \n", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode user info into struct
	var googleUser models.GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		fmt.Println("Error decoding use info: ", err)
		http.Error(w, "Could not parse user info", http.StatusInternalServerError)
		return
	}

	fmt.Printf("User Info: %+v\n", googleUser)

	// store/update user info in db
	userID, err := database.StoreUserInDB(googleUser, "google")
	if err != nil {
		fmt.Printf("Error storing user in DB: %v\n", err)
		http.Error(w, "Failed to store user info", http.StatusInternalServerError)
		return
	}

	// generate session or jwt for persistent login
	tokenString, err := generateJWT(userID)
	if err != nil {
		fmt.Printf("Error storing user in DB: %v\n", err)
		http.Error(w, "Failed to store user info", http.StatusInternalServerError)
		return
	}

	//token == cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
	//redirect user to homepage
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func GoogleLogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear session or token (this depends on how you’re managing sessions)
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Unix(0, 0), // Expire the cookie immediately
		Path:    "/",
	})

	// Redirect to homepage or login page after logout
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
