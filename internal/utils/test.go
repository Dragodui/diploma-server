package utils

import "context"

type contextKey string

const UserIDKey contextKey = "userID"

func WithUserID(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, UserIDKey, id)
}
