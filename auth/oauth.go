// follow this https://permify.co/post/implement-oauth-2-golang-app/#key-components-of-oauth-2-0-authentication-system
package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/haikali3/gymbara-backend/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleOauthConfig = &oauth2.Config{
	RedirectURL: "http://localhost:8080/oauth/callback",
	Scopes:      []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:    google.Endpoint,
}

func GenerateStateOAuthCookie(w http.ResponseWriter) string {
	b := make([]byte, 16)
	rand.Read(b)
	oauthStateString := base64.URLEncoding.EncodeToString(b)

	fmt.Println("Generated OAuth state:", oauthStateString)

	http.SetCookie(w, &http.Cookie{
		Name:    "oauthstate",
		Value:   oauthStateString,
		Expires: time.Now().Add(24 * time.Hour),
		Path:    "/", //enable cookie avalable across all paths
	})

	return oauthStateString
}

// Redirects user to Google login
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {

	// load from .env
	GoogleOauthConfig.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleOauthConfig.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")

	// Print ClientID and ClientSecret for verification
	fmt.Println("Google Client ID (Handler):", GoogleOauthConfig.ClientID)
	fmt.Println("Google Client Secret (Handler):", GoogleOauthConfig.ClientSecret)

	// generate state, set it in cookie, add it to AuthCodeURL
	oauthStateString := GenerateStateOAuthCookie(w)
	url := GoogleOauthConfig.AuthCodeURL(oauthStateString)
	fmt.Println("Redirecting to Google OAuth URL:", url)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Handles Google OAuth callback and processes user info
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	//load ClientId and ClientSecret from env
	GoogleOauthConfig.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleOauthConfig.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")

	//check .env
	fmt.Println("Google Client ID:", GoogleOauthConfig.ClientID)
	fmt.Println("Google Client Secret:", GoogleOauthConfig.ClientSecret)

	//get state from cookie and verify
	stateCookie, err := r.Cookie("oauthstate")
	if err != nil || r.FormValue("state") != stateCookie.Value {
		fmt.Println("Error retrieving state cookie:", err)
		http.Error(w, "Invalid OAuth state cookie", http.StatusBadRequest)
		return
	}
	fmt.Println("OAuth State received:", stateCookie.Value)
	fmt.Println("OAuth State from request:", r.FormValue("state"))

	//verify state in query(url) == state in cookie
	if r.FormValue("state") != stateCookie.Value {
		fmt.Println("State mismatch: expected ", stateCookie.Value, "got ", r.FormValue("state"))
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	//exchange auth code for access token
	token, err := GoogleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		fmt.Println("Error exchanging code for token:", err)
		http.Error(w, "Could not get token", http.StatusInternalServerError)
		return
	}
	fmt.Println("Access Token:", token.AccessToken)

	// get user info from google
	client := GoogleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		fmt.Println("Error fetching user info:", err)
		http.Error(w, "Could not get token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode user info
	var googleUser models.GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		fmt.Println("Error decoding use info")
		http.Error(w, "Could not parse user info", http.StatusInternalServerError)
		return
	}

	fmt.Printf("User Info: %+v\n", googleUser)

	// Option 1: Display user info directly in the response (temporary)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Welcome, %s!</h1><p>Email: %s</p><p>Picture:</p><img src='%s' alt='User picture'>", googleUser.Name, googleUser.Email, googleUser.Picture)

	//TODO: replace 2 second delay with storing user data in a database and redirect to gymbara.com/
	// Option 2: Redirect to homepage after a short delay
	time.Sleep(2 * time.Second) // Show user info for 2 seconds
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
	// TODO: add logout page later
}
