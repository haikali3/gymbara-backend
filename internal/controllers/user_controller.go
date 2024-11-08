package controllers

import (
	"encoding/json"
	"net/http"
)

func GetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	authCookie, err := r.Cookie("access_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken := authCookie.Value
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch user details from Google", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Error decoding user info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}
