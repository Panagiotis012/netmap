package middleware

import (
	"context"
	"net/http"
)

type SessionValidator interface {
	Validate(ctx context.Context, token string) (bool, error)
}

type SetupChecker interface {
	IsSetup(ctx context.Context) bool
}

// RequireAuth checks session cookie. If auth is not yet configured (no password),
// it passes through to allow first-run setup.
func RequireAuth(sessions SessionValidator, setup SetupChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If no password has been set yet, allow all requests through.
			if !setup.IsSetup(r.Context()) {
				next.ServeHTTP(w, r)
				return
			}
			cookie, err := r.Cookie("nm_session")
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			ok, err := sessions.Validate(r.Context(), cookie.Value)
			if err != nil || !ok {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
