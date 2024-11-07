package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/haikali3/gymbara-backend/models"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func generateJWT(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		// expire after 30 days
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	return token.SignedString(jwtKey)
}

func generateJWTWithUserDetails(user models.GoogleUser) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      user.ID,
		"email":   user.Email,
		"name":    user.Name,
		"picture": user.Picture,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(), // Token expires in 30 days
	})
	return token.SignedString(jwtKey)
}
