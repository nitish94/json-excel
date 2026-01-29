package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID creates a random 8-byte hex string
func GenerateID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "default"
	}
	return hex.EncodeToString(bytes)
}
