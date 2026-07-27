package apikey_test

import (
	"strings"
	"testing"

	"consolehub/internal/apikey"
)

func TestAPIKey_GenerateAndVerify(t *testing.T) {
	key, err := apikey.Generate()
	if err != nil {
		t.Fatalf("failed to generate API key: %v", err)
	}

	if !strings.HasPrefix(key, "sk_") {
		t.Errorf("expected key to start with 'sk_', got %s", key)
	}

	if !strings.Contains(key, "_crc32_") {
		t.Errorf("expected key to contain '_crc32_', got %s", key)
	}

	if !apikey.Verify(key) {
		t.Errorf("expected generated key %s to pass verification", key)
	}
}

func TestAPIKey_TamperedKeyFailsVerify(t *testing.T) {
	key, err := apikey.Generate()
	if err != nil {
		t.Fatalf("failed to generate API key: %v", err)
	}

	// Tamper with payload
	parts := strings.Split(key, "_crc32_")
	tamperedPayload := parts[0] + "X" + "_crc32_" + parts[1]
	if apikey.Verify(tamperedPayload) {
		t.Errorf("expected tampered key %s to fail verification", tamperedPayload)
	}

	// Tamper with checksum
	tamperedChecksum := parts[0] + "_crc32_" + parts[1] + "0"
	if apikey.Verify(tamperedChecksum) {
		t.Errorf("expected key with tampered checksum %s to fail verification", tamperedChecksum)
	}

	// Invalid prefix
	invalidPrefix := "pk_" + strings.TrimPrefix(key, "sk_")
	if apikey.Verify(invalidPrefix) {
		t.Errorf("expected key with invalid prefix %s to fail verification", invalidPrefix)
	}
}

func TestAPIKey_Base62EncodingDecoding(t *testing.T) {
	input := []byte("hello world 1234")
	encoded := apikey.EncodeBase62(input)
	decoded, err := apikey.DecodeBase62(encoded)
	if err != nil {
		t.Fatalf("failed to decode base62: %v", err)
	}

	if string(decoded) != string(input) {
		t.Errorf("expected %s, got %s", string(input), string(decoded))
	}
}
