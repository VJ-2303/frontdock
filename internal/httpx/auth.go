package httpx

import (
	"context"
	"net/http"
	"strings"

	"github.com/VJ-2303/frontdock/internal/auth"
	"github.com/google/uuid"
)

type userCtxKey struct{}

type AuthUser struct {
	ID            uuid.UUID
	EmailVerified bool
}

func RequireAuth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			raw, ok := strings.CutPrefix(h, "Bearer ")
			if !ok || raw == "" {
				Error(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
				return
			}
			claims, err := auth.ParseToken(secret, raw)
			if err != nil {
				Error(w, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
				return
			}
			uid, err := uuid.Parse(claims.Subject)
			if err != nil {
				Error(w, http.StatusUnauthorized, "unauthorized", "malformed token subject")
				return
			}
			ctx := context.WithValue(r.Context(), userCtxKey{}, AuthUser{
				ID: uid, EmailVerified: claims.EmailVerified,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireVerified(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFrom(r.Context())
		if !ok || !u.EmailVerified {
			Error(w, http.StatusForbidden, "email_not_verified", "verify your email address before using this endpoint")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func UserFrom(ctx context.Context) (AuthUser, bool) {
	u, ok := ctx.Value(userCtxKey{}).(AuthUser)
	return u, ok
}
