package core

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// Fallback to SHA256 if bcrypt fails
		hash := sha256.Sum256([]byte(password))
		return hex.EncodeToString(hash[:])
	}
	return string(bytes)
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(password, hashedPassword string) bool {
	// Try bcrypt first
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == nil {
		return true
	}
	
	// Fallback to SHA256 comparison for legacy hashes
	hash := sha256.Sum256([]byte(password))
	expectedHash := hex.EncodeToString(hash[:])
	return hashedPassword == expectedHash
}
