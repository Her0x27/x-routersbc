package utils

import (
        "crypto/rand"
        "encoding/hex"
        "fmt"
)



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
