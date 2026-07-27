package ulid

import (
	"crypto/rand"
	"sync"
	"time"
)

const crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

var (
	entropyMu sync.Mutex
)

// Make generates a new canonical 26-character Crockford Base32 ULID.
func Make() string {
	entropyMu.Lock()
	defer entropyMu.Unlock()

	nowMs := uint64(time.Now().UnixMilli())

	var dst [26]byte

	// 48-bit Timestamp (10 chars)
	dst[0] = crockfordAlphabet[(nowMs>>45)&0x1F]
	dst[1] = crockfordAlphabet[(nowMs>>40)&0x1F]
	dst[2] = crockfordAlphabet[(nowMs>>35)&0x1F]
	dst[3] = crockfordAlphabet[(nowMs>>30)&0x1F]
	dst[4] = crockfordAlphabet[(nowMs>>25)&0x1F]
	dst[5] = crockfordAlphabet[(nowMs>>20)&0x1F]
	dst[6] = crockfordAlphabet[(nowMs>>15)&0x1F]
	dst[7] = crockfordAlphabet[(nowMs>>10)&0x1F]
	dst[8] = crockfordAlphabet[(nowMs>>5)&0x1F]
	dst[9] = crockfordAlphabet[nowMs&0x1F]

	// 80-bit Entropy (10 random bytes -> 16 chars)
	var entropy [10]byte
	_, _ = rand.Read(entropy[:])

	dst[10] = crockfordAlphabet[(entropy[0]>>3)&0x1F]
	dst[11] = crockfordAlphabet[((entropy[0]&0x07)<<2)|((entropy[1]>>6)&0x03)]
	dst[12] = crockfordAlphabet[(entropy[1]>>1)&0x1F]
	dst[13] = crockfordAlphabet[((entropy[1]&0x01)<<4)|((entropy[2]>>4)&0x0F)]
	dst[14] = crockfordAlphabet[((entropy[2]&0x0F)<<1)|((entropy[3]>>7)&0x01)]
	dst[15] = crockfordAlphabet[(entropy[3]>>2)&0x1F]
	dst[16] = crockfordAlphabet[((entropy[3]&0x03)<<3)|((entropy[4]>>5)&0x07)]
	dst[17] = crockfordAlphabet[entropy[4]&0x1F]
	dst[18] = crockfordAlphabet[(entropy[5]>>3)&0x1F]
	dst[19] = crockfordAlphabet[((entropy[5]&0x07)<<2)|((entropy[6]>>6)&0x03)]
	dst[20] = crockfordAlphabet[(entropy[6]>>1)&0x1F]
	dst[21] = crockfordAlphabet[((entropy[6]&0x01)<<4)|((entropy[7]>>4)&0x0F)]
	dst[22] = crockfordAlphabet[((entropy[7]&0x0F)<<1)|((entropy[8]>>7)&0x01)]
	dst[23] = crockfordAlphabet[(entropy[8]>>2)&0x1F]
	dst[24] = crockfordAlphabet[((entropy[8]&0x03)<<3)|((entropy[9]>>5)&0x07)]
	dst[25] = crockfordAlphabet[entropy[9]&0x1F]

	return string(dst[:])
}

// IsValid checks if a string is a valid 26-character Crockford Base32 ULID.
func IsValid(id string) bool {
	if len(id) != 26 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		valid := false
		for j := 0; j < len(crockfordAlphabet); j++ {
			if c == crockfordAlphabet[j] {
				valid = true
				break
			}
		}
		if !valid {
			return false
		}
	}
	return true
}
