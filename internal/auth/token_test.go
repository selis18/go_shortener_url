package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/selis18/go_shortener_url/internal/config"
)

func TestTokenRoundTrip(t *testing.T) {
	token, err := BuildToken("alice")
	if err != nil {
		t.Fatal(err)
	}
	userID, err := ParseToken(token)
	if err != nil || userID != "alice" {
		t.Fatalf("ParseToken() = %q, %v; want alice, nil", userID, err)
	}
}

func TestParseTokenRejectsInvalidTokens(t *testing.T) {
	sign := func(method jwt.SigningMethod, expires time.Time, key interface{}) string {
		t.Helper()
		token, err := jwt.NewWithClaims(method, claims{
			RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expires)},
			UserID:           "alice",
		}).SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	future := time.Now().Add(time.Hour)
	for name, token := range map[string]string{
		"empty":           "",
		"malformed":       "hello",
		"expired":         sign(jwt.SigningMethodHS256, time.Now().Add(-time.Hour), []byte(config.GetSecretKey())),
		"wrong signature": sign(jwt.SigningMethodHS256, future, []byte(config.GetSecretKey()+"wrong")),
		"unsigned":        sign(jwt.SigningMethodNone, future, jwt.UnsafeAllowNoneSignatureType),
	} {
		t.Run(name, func(t *testing.T) {
			userID, err := ParseToken(token)
			if err == nil || userID != "" {
				t.Fatalf("ParseToken() = %q, %v; want empty user ID and error", userID, err)
			}
		})
	}
}
