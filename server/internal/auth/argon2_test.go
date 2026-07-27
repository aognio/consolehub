package auth_test

import (
	"testing"

	"consolehub/internal/auth"
)

func TestArgon2_HashAndVerify(t *testing.T) {
	rawKey := "sk_3q2Z7x9P2R8LmNkJd4Hf6Y_crc32_2AB9XQ"
	hash, err := auth.HashArgon2(rawKey)
	if err != nil {
		t.Fatalf("failed to hash key: %v", err)
	}

	valid, err := auth.VerifyArgon2(rawKey, hash)
	if err != nil || !valid {
		t.Fatalf("failed to verify valid key against argon2 hash: %v", err)
	}

	invalidKey := "sk_3q2Z7x9P2R8LmNkJd4Hf6Y_crc32_WRONG1"
	validWrong, err := auth.VerifyArgon2(invalidKey, hash)
	if err == nil && validWrong {
		t.Fatalf("expected wrong key to fail verification")
	}
}
