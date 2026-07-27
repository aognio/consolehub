package middleware

import (
	"context"
	"net/http"
	"time"

	"consolehub/internal/auth"
)

type contextKey string

const (
	UserKey   contextKey = "user_claims"
	TenantKey contextKey = "tenant_id"
)

type Middleware struct {
	cookieSecret string
}

func New(cookieSecret string) *Middleware {
	return &Middleware{
		cookieSecret: cookieSecret,
	}
}

// RequireAuth extracts session cookie, verifies HMAC token, and injects claims into request context.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("consolehub_session")
		if err != nil || cookie.Value == "" {
			if isHTMX(r) {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		claims, err := auth.ValidateSessionToken(cookie.Value, m.cookieSecret)
		if err != nil {
			// Clear invalid cookie
			http.SetCookie(w, &http.Cookie{
				Name:    "consolehub_session",
				Value:   "",
				Path:    "/",
				Expires: time.Unix(0, 0),
			})
			if isHTMX(r) {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserClaims retrieves active session claims from request context.
func GetUserClaims(r *http.Request) *auth.SessionClaims {
	claims, _ := r.Context().Value(UserKey).(*auth.SessionClaims)
	return claims
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
