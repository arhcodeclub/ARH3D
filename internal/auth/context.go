package auth

import (
	"context"

	"github.com/arhcodeclub/arh3d/internal/models"
)

type contextKey string

const userKey = contextKey("user")

func ContextWithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context) *models.User {
	val := ctx.Value(userKey)
	if u, ok := val.(*models.User); ok {
		return u
	}

	return nil
}
