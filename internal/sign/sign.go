package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const HashHeader = "HashSHA256"

// Hash calculates the HMAC-SHA256 hash of the given data using the provided key.
func Hash(data []byte, key string) string {
	// Create a new HMAC-SHA256 hash.
	h := hmac.New(sha256.New, []byte(key))
	// Write the data to the hash.
	h.Write(data)
	// Sum the hash and return the result.
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}

// Verify checks if the given data matches the HMAC-SHA256 hash of the data using the provided key.
func Verify(data []byte, key string, headerHash string) (bool, error) {
	// Create a new HMAC-SHA256 hash.
	h := hmac.New(sha256.New, []byte(key))
	// Write the data to the hash.
	h.Write(data)
	// Sum the hash and return the result.
	hash := h.Sum(nil)
	headerHashBytes, err := hex.DecodeString(headerHash)
	if err != nil {
		return false, fmt.Errorf("failed to decode header hash: %w", err)
	}

	// Compare the hash with the data.
	verified := hmac.Equal(hash, headerHashBytes)
	if !verified {
		return false, fmt.Errorf("hash verification failed: provided=%s want=%s", headerHash, hex.EncodeToString(hash))
	}

	return verified, nil
}
