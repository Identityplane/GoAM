package oauth2

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"regexp"
)

const (
	// MinCodeVerifierLength is the minimum length of a code verifier
	MinCodeVerifierLength = 43
	// MaxCodeVerifierLength is the maximum length of a code verifier
	MaxCodeVerifierLength = 128
)

var (
	// ErrInvalidCodeVerifierLength is returned when the code verifier length is invalid
	ErrInvalidCodeVerifierLength = errors.New("code verifier length must be between 43 and 128 characters")
	// ErrInvalidCodeVerifierFormat is returned when the code verifier format is invalid
	ErrInvalidCodeVerifierFormat = errors.New("code verifier must only contain characters in the range [A-Za-z0-9-._~]")
)

// ValidateCodeVerifier validates a code verifier according to RFC 7636
func ValidateCodeVerifier(codeVerifier string) error {
	// Check length
	if len(codeVerifier) < MinCodeVerifierLength || len(codeVerifier) > MaxCodeVerifierLength {
		return ErrInvalidCodeVerifierLength
	}

	// Check format
	validFormat := regexp.MustCompile(`^[A-Za-z0-9-._~]+$`)
	if !validFormat.MatchString(codeVerifier) {
		return ErrInvalidCodeVerifierFormat
	}

	return nil
}

// GenerateCodeChallenge generates a code challenge from a code verifier using SHA-256
func GenerateCodeChallenge(codeVerifier string) (string, error) {
	// Validate the code verifier
	if err := ValidateCodeVerifier(codeVerifier); err != nil {
		return "", err
	}

	// Calculate SHA-256 hash
	hash := sha256.Sum256([]byte(codeVerifier))

	// Base64 URL-safe encode without padding
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return challenge, nil
}

// VerifyCodeChallenge verifies that a code verifier matches a code challenge
func VerifyCodeChallenge(codeVerifier, codeChallenge string) (bool, error) {
	// Generate the challenge from the verifier
	generatedChallenge, err := GenerateCodeChallenge(codeVerifier)
	if err != nil {
		return false, err
	}

	// Compare the generated challenge with the provided challenge
	return generatedChallenge == codeChallenge, nil
}
