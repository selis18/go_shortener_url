package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/selis18/go_shortener_url/internal/config"
)

type claims struct {
	jwt.RegisteredClaims
	UserID string
}

func BuildToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(3 * time.Hour)),
		},
		UserID: userID,
	})
	return token.SignedString([]byte(config.GetSecretKey()))
}

func ParseToken(tokenString string) (string, error) {
	c := &claims{}
	token, err := jwt.ParseWithClaims(tokenString, c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(config.GetSecretKey()), nil
	})
	if err != nil {
		return "", err
	}
	if token == nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	return c.UserID, nil
}
