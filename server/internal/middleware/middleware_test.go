package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"consolehub/internal/auth"
	"consolehub/internal/middleware"
	"consolehub/internal/models"
)

func TestAuthMiddleware_Unauthorized(t *testing.T) {
	secret := "secret-32-bytes-long-key-for-test"
	mw := middleware.New(secret)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.RequireAuth(next)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther && rec.Code != http.StatusUnauthorized {
		t.Errorf("expected redirect (303) or unauthorized (401), got %d", rec.Code)
	}
}

func TestAuthMiddleware_Authorized(t *testing.T) {
	secret := "secret-32-bytes-long-key-for-test"
	mw := middleware.New(secret)

	user := &models.User{
		ID:    "usr-1",
		Email: "admin@example.com",
		Role:  models.RoleSuperAdmin,
	}
	token, _ := auth.CreateSessionToken(user, secret, 1*time.Hour)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.GetUserClaims(r)
		if claims == nil || claims.UserID != user.ID {
			t.Errorf("expected claims UserID %s, got %v", user.ID, claims)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.RequireAuth(next)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	req.AddCookie(&http.Cookie{Name: "consolehub_session", Value: token})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected OK 200, got %d", rec.Code)
	}
}
