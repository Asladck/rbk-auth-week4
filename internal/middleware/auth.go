package middleware

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"weather-auth-api/internal/domain"
	"weather-auth-api/pkg/apperrors"
	"weather-auth-api/pkg/jwtutil"
)

func Auth(jwtManager *jwtutil.Manager, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, err := extractBearerToken(r)
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, "missing or malformed Authorization header")
				return
			}

			claims, err := jwtManager.Parse(tokenStr)
			if err != nil {
				log.Debug("invalid token", zap.Error(err), zap.String("path", r.URL.Path))
				writeAuthError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := domain.ContextWithUser(r.Context(), domain.AuthUser{
				ID:    claims.UserID,
				Email: claims.Email,
				Role:  claims.Role,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := domain.UserFromContext(r.Context())
			if !ok {
				writeAuthError(w, http.StatusUnauthorized, "not authenticated")
				return
			}

			if _, permitted := allowed[user.Role]; !permitted {
				writeAuthError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractBearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", apperrors.ErrUnauthorized
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", apperrors.ErrUnauthorized
	}
	return strings.TrimSpace(parts[1]), nil
}

func writeAuthError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}
