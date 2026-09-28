package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/selis18/go_shortener_url/internal/logger"
	"go.uber.org/zap"
)

func CookieMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string
		cookies, err := r.Cookie("userId")
		if err == nil {
			userID, err = ParseToken(cookies.Value)
		}

		if err != nil {
			userID, err = createCookie(w)
			if err != nil {
				logger.Log.Error("create user cookie", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		ctx := WithUserID(r.Context(), userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func makeHexUserID() (string, error) {
	userID := make([]byte, 16)
	_, err := rand.Read(userID)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(userID), nil
}

func createCookie(w http.ResponseWriter) (string, error) {
	userID, err := makeHexUserID()
	if err != nil {
		return "", err
	}
	tokenString, err := BuildToken(userID)
	if err != nil {
		return "", err
	}
	cookie := &http.Cookie{
		Name:     "userId",
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
	}

	http.SetCookie(w, cookie)

	return userID, nil
}
