package auth

import "context"

type UserID struct{}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserID{}, userID)
}

func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserID{}).(string)
	return id, ok && id != ""
}
