package main

import (
	"context"

	"github.com/go-dew/dew"
)

type userCtxKey struct{}

type CurrentUser struct {
	ID int
}

func ctxWithCurrUser(ctx context.Context, u *CurrentUser) context.Context {
	return context.WithValue(ctx, userCtxKey{}, u)
}

func currUserFromCtx(ctx context.Context) *CurrentUser {
	return ctx.Value(userCtxKey{}).(*CurrentUser)
}

// isAuthorized checks if the current user is authorized.
func isAuthorized(ctx context.Context) bool {
	return currUserFromCtx(ctx).ID == AdminID
}

func AdminOnly(next dew.Middleware) dew.Middleware {
	return dew.MiddlewareFunc(func(ctx dew.Context) error {
		if !isAuthorized(ctx.Context()) {
			// Return an unauthorized error.
			return ErrUnauthorized
		}
		// Continue to the next middleware or handler.
		return next.Handle(ctx)
	})
}
