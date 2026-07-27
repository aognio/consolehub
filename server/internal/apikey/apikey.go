package apikey

import (
	"crypto/rand"
	"fmt"
	"hash/crc32"
	"math/big"
	"strings"
)

const (
	Prefix    = "sk_"
	Separator = "_crc32_"
	Alphabet  = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var base62Radix = big.NewInt(62)

// EncodeBase62 encodes a byte slice into a Base62 string.
func EncodeBase62(b []byte) string {
	if len(b) == 0 {
		return ""
	}

	var num big.Int
	num.SetBytes(b)

	if num.Cmp(big.NewInt(0)) == 0 {
		return "0"
	}

	var chars []byte
	var zero big.Int
	var mod big.Int

	for num.Cmp(&zero) > 0 {
		num.DivMod(&num, base62Radix, &mod)
		chars = append(chars, Alphabet[mod.Int64()])
	}

	// Reverse string
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}

	return string(chars)
}

// DecodeBase62 decodes a Base62 string back into a byte slice.
func DecodeBase62(s string) ([]byte, error) {
	if s == "" {
		return nil, nil
	}

	var num big.Int
	for i := 0; i < len(s); i++ {
		idx := strings.IndexByte(Alphabet, s[i])
		if idx < 0 {
			return nil, fmt.Errorf("invalid base62 character '%c'", s[i])
		}
		num.Mul(&num, base62Radix)
		num.Add(&num, big.NewInt(int64(idx)))
	}

	return num.Bytes(), nil
}

// Generate creates a new API key string in the format:
// sk_<base62(16_random_bytes)>_crc32_<base62(crc32_checksum)>
func Generate() (string, error) {
	entropy := make([]byte, 16)
	if _, err := rand.Read(entropy); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	payloadBase62 := EncodeBase62(entropy)

	// Calculate 4-byte CRC32 checksum over the 16 random bytes
	checksumUint := crc32.ChecksumIEEE(entropy)
	checksumBytes := []byte{
		byte(checksumUint >> 24),
		byte(checksumUint >> 16),
		byte(checksumUint >> 8),
		byte(checksumUint),
	}
	checksumBase62 := EncodeBase62(checksumBytes)

	return fmt.Sprintf("%s%s%s%s", Prefix, payloadBase62, Separator, checksumBase62), nil
}

// MustGenerate calls Generate and panics if an error occurs.
func MustGenerate() string {
	key, err := Generate()
	if err != nil {
		panic(err)
	}
	return key
}

// Verify validates the format, Base62 decoding, and CRC32 checksum of an API key.
func Verify(key string) bool {
	if !strings.HasPrefix(key, Prefix) {
		return false
	}

	trimmed := strings.TrimPrefix(key, Prefix)
	parts := strings.Split(trimmed, Separator)
	if len(parts) != 2 {
		return false
	}

	payloadBase62 := parts[0]
	checksumBase62 := parts[1]

	if payloadBase62 == "" || checksumBase62 == "" {
		return false
	}

	entropy, err := DecodeBase62(payloadBase62)
	if err != nil || len(entropy) == 0 {
		return false
	}

	checksumUint := crc32.ChecksumIEEE(entropy)
	expectedChecksumBytes := []byte{
		byte(checksumUint >> 24),
		byte(checksumUint >> 16),
		byte(checksumUint >> 8),
		byte(checksumUint),
	}
	expectedChecksumBase62 := EncodeBase62(expectedChecksumBytes)

	return checksumBase62 == expectedChecksumBase62
}
