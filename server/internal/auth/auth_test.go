package auth_test

import (
	"testing"
	"time"

	"consolehub/internal/auth"
	"consolehub/internal/models"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecretPassw0rd!"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !auth.CheckPasswordHash(password, hash) {
		t.Errorf("expected password verification to succeed")
	}

	if auth.CheckPasswordHash("WrongPassword", hash) {
		t.Errorf("expected incorrect password verification to fail")
	}
}

func TestSessionCookieToken(t *testing.T) {
	secret := "my-super-secret-key-32-bytes-long!!"
	user := &models.User{
		ID:    "usr-12345",
		Email: "admin@example.com",
		Role:  models.RoleSuperAdmin,
	}
	duration := 1 * time.Hour

	token, err := auth.CreateSessionToken(user, secret, duration)
	if err != nil {
		t.Fatalf("failed to create session token: %v", err)
	}

	claims, err := auth.ValidateSessionToken(token, secret)
	if err != nil {
		t.Fatalf("failed to validate session token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("expected UserID %s, got %s", user.ID, claims.UserID)
	}
	if claims.Email != user.Email {
		t.Errorf("expected Email %s, got %s", user.Email, claims.Email)
	}
	if claims.Role != user.Role {
		t.Errorf("expected Role %s, got %s", user.Role, claims.Role)
	}
}

func TestExpiredSessionCookieToken(t *testing.T) {
	secret := "my-super-secret-key-32-bytes-long!!"
	user := &models.User{
		ID:    "usr-12345",
		Email: "admin@example.com",
		Role:  models.RoleSuperAdmin,
	}

	// Negative duration = already expired token
	token, err := auth.CreateSessionToken(user, secret, -1*time.Minute)
	if err != nil {
		t.Fatalf("failed to create expired session token: %v", err)
	}

	_, err = auth.ValidateSessionToken(token, secret)
	if err == nil {
		t.Fatal("expected validation error for expired token, got nil")
	}
}
