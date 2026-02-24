package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64
	Roles  []string
}

func GenerateJWT(userID int64, roles []string, secret string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"roles":   roles,
		"exp":     time.Now().Add(duration).Unix(),
	}
	toke := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return toke.SignedString([]byte(secret))
}

func ValidateJWT(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claimsMap := token.Claims.(jwt.MapClaims)

	userID := int64(claimsMap["user_id"].(float64))
	rolesInterface := claimsMap["roles"].([]interface{})

	roles := make([]string, len(rolesInterface))
	for i, v := range rolesInterface {
		roles[i] = v.(string)
	}

	return &Claims{
		UserID: userID,
		Roles:  roles,
	}, nil
}
