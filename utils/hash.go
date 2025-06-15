package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

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

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// GenerateSessionToken generates a secure session token
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session token: %v", err)
	}
	return hex.EncodeToString(bytes), nil
}
