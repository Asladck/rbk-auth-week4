package domain

import "context"

type contextKey string

const (
	contextKeyUser contextKey = "auth_user"
)

type AuthUser struct {
	ID    string
	Email string
	Role  string
}

func ContextWithUser(ctx context.Context, u AuthUser) context.Context {
	return context.WithValue(ctx, contextKeyUser, u)
}

func UserFromContext(ctx context.Context) (AuthUser, bool) {
	u, ok := ctx.Value(contextKeyUser).(AuthUser)
	return u, ok
}

func MustUserFromContext(ctx context.Context) AuthUser {
	u, ok := UserFromContext(ctx)
	if !ok {
		panic("MustUserFromContext: no auth user in context — is AuthMiddleware applied?")
	}
	return u
}
