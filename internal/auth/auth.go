package auth

import (
	"context"
	"errors"
)

var (
	ErrUserIDNotFound    = errors.New("user ID not found in context")
	ErrUserIDInvalidType = errors.New("user ID has unexpected type")
	ErrUserIDEmpty       = errors.New("user ID is empty")
)

type userIDKey struct{}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func GetUserID(ctx context.Context) (string, error) {
	value := ctx.Value(userIDKey{})
	if value == nil {
		return "", ErrUserIDNotFound
	}
	id, ok := value.(string)
	if !ok {
		return "", ErrUserIDInvalidType
	}
	if id == "" {
		return "", ErrUserIDEmpty
	}
	return id, nil
}
