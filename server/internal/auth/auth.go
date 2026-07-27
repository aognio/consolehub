package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"consolehub/internal/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidToken = errors.New("invalid or tampered session token")
	ErrExpiredToken = errors.New("session token has expired")
)

// SessionClaims represents the payload encoded inside signed session tokens.
type SessionClaims struct {
	UserID    string    `json:"uid"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"exp"`
}

// HashPassword generates a bcrypt hash of the plain-text password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a plain-text password against a bcrypt hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CreateSessionToken generates an HMAC-SHA256 signed session token.
func CreateSessionToken(user *models.User, secret string, duration time.Duration) (string, error) {
	claims := SessionClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		ExpiresAt: time.Now().Add(duration),
	}

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal session claims: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signature := computeHMAC(encodedPayload, secret)

	return fmt.Sprintf("%s.%s", encodedPayload, signature), nil
}

// ValidateSessionToken verifies the HMAC-SHA256 signature and expiration of a session token.
func ValidateSessionToken(token, secret string) (*SessionClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}

	encodedPayload, signature := parts[0], parts[1]
	expectedSig := computeHMAC(encodedPayload, secret)

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return nil, ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims SessionClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().After(claims.ExpiresAt) {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

func computeHMAC(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
