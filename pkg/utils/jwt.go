package utils

import (
	"errors"
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
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil || !token.Valid {
		return nil, err
	}

	claimsMap, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// exp is mandatory: a token without expiration must never be accepted.
	expRaw, ok := claimsMap["exp"].(float64)
	if !ok {
		return nil, errors.New("token missing expiration")
	}
	if time.Now().Unix() >= int64(expRaw) {
		return nil, errors.New("token expired")
	}

	userID, ok := claimsMap["user_id"].(float64)
	if !ok {
		return nil, errors.New("token missing user_id")
	}

	roles := []string{}
	if rolesInterface, ok := claimsMap["roles"].([]interface{}); ok {
		for _, v := range rolesInterface {
			if role, ok := v.(string); ok {
				roles = append(roles, role)
			}
		}
	}

	return &Claims{
		UserID: int64(userID),
		Roles:  roles,
	}, nil
}
