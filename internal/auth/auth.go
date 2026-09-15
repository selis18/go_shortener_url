package auth

import "context"

type UserId struct{}

func WithUserId(ctx context.Context, userId string) context.Context {
	return context.WithValue(ctx, UserId{}, userId)
}

func GetUserId(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserId{}).(string)
	return id, ok && id != ""
}
