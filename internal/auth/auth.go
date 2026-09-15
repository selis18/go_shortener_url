package auth

import "context"

type UserID struct{}

func WithUserID(ctx context.Context, userId string) context.Context {
	return context.WithValue(ctx, UserID{}, userId)
}

func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserID{}).(string)
	return id, ok && id != ""
}
