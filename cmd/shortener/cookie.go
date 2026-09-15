package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/selis18/go_shortener_url/internal/auth"
	"github.com/selis18/go_shortener_url/internal/config"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func cookieMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var id string
		claims := &Claims{}
		cookies, err := r.Cookie("userId")
		if err != nil {
			log.Println(err)
		}
		if cookies == nil {
			cookies, err = createCookie(claims, w)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		token, err := jwt.ParseWithClaims(cookies.Value, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(config.GetSecretKey()), nil
		})

		if err != nil || token == nil || !token.Valid {
			cookies, err = createCookie(claims, w)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		id = claims.UserID
		ctx := auth.WithUserId(r.Context(), id)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func makeHexUserId() (string, error) {
	userId := make([]byte, 16)
	_, err := rand.Read(userId)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(userId), nil
}

func createCookie(c *Claims, w http.ResponseWriter) (*http.Cookie, error) {
	userId, err := makeHexUserId()
	if err != nil {
		return nil, err
	}
	c.UserID = userId
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(3 * time.Hour)),
		},
		UserID: userId,
	})
	tokenString, err := token.SignedString([]byte(config.GetSecretKey()))
	if err != nil {
		return nil, err
	}
	cookie := &http.Cookie{
		Name:     "userId",
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
	}

	http.SetCookie(w, cookie)

	return cookie, nil
}
