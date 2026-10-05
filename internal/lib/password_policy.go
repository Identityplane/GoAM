package lib

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	DefaultMinPasswordLength = 8
	DefaultMaxPasswordLength = 64

	configMinLength = "minLength"
	configMaxLength = "maxLength"
)

// PasswordLimits is the min/max character count for a new password.
// Zero Min means no minimum besides non-empty. Zero Max means no maximum.
type PasswordLimits struct {
	Min int
	Max int
}

// NormalizeUsername trims surrounding whitespace from a username.
func NormalizeUsername(username string) string {
	return strings.TrimSpace(username)
}

// ValidateUsername rejects usernames that are empty after trimming.
func ValidateUsername(username string) error {
	if NormalizeUsername(username) == "" {
		return fmt.Errorf("Username is required")
	}
	return nil
}

// PasswordLimitsForAsk reads minLength/maxLength for askPassword.
// Unset values are 0 so login can still accept short or long legacy passwords.
func PasswordLimitsForAsk(customConfig map[string]string) PasswordLimits {
	return PasswordLimits{
		Min: parseLength(customConfig, configMinLength, 0),
		Max: parseLength(customConfig, configMaxLength, 0),
	}
}

// PasswordLimitsForStore reads minLength/maxLength when hashing a new password.
// Defaults to DefaultMinPasswordLength and DefaultMaxPasswordLength when unset.
func PasswordLimitsForStore(customConfig map[string]string) PasswordLimits {
	return PasswordLimits{
		Min: parseLength(customConfig, configMinLength, DefaultMinPasswordLength),
		Max: parseLength(customConfig, configMaxLength, DefaultMaxPasswordLength),
	}
}

func parseLength(customConfig map[string]string, key string, defaultLen int) int {
	if customConfig == nil {
		return defaultLen
	}
	raw := strings.TrimSpace(customConfig[key])
	if raw == "" {
		return defaultLen
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return defaultLen
	}
	return n
}

// ValidateNewPassword rejects an empty password, then min/max when those limits are > 0.
// Passwords are not trimmed (spaces are significant).
func ValidateNewPassword(password string, limits PasswordLimits) error {
	if password == "" {
		return fmt.Errorf("Password is required")
	}
	n := utf8.RuneCountInString(password)
	if limits.Min > 0 && n < limits.Min {
		return fmt.Errorf("Password must be at least %d characters", limits.Min)
	}
	if limits.Max > 0 && n > limits.Max {
		return fmt.Errorf("Password must be at most %d characters", limits.Max)
	}
	return nil
}
