package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID int32, secret string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
	}
	toke := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return toke.SignedString([]byte(secret))
}
