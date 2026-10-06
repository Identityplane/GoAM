package lib

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeUsername(t *testing.T) {
	assert.Equal(t, "alice", NormalizeUsername("alice"))
	assert.Equal(t, "alice", NormalizeUsername("  alice  "))
	assert.Equal(t, "", NormalizeUsername("   "))
	assert.Equal(t, "", NormalizeUsername(""))
}

func TestValidateUsername(t *testing.T) {
	assert.NoError(t, ValidateUsername("alice"))
	assert.NoError(t, ValidateUsername("  alice  "))
	assert.EqualError(t, ValidateUsername(""), "Username is required")
	assert.EqualError(t, ValidateUsername("   "), "Username is required")
}

func TestPasswordLimitsForAsk(t *testing.T) {
	assert.Equal(t, PasswordLimits{}, PasswordLimitsForAsk(nil))
	assert.Equal(t, PasswordLimits{}, PasswordLimitsForAsk(map[string]string{}))
	assert.Equal(t, PasswordLimits{Min: 8, Max: 64}, PasswordLimitsForAsk(map[string]string{
		"minLength": "8",
		"maxLength": "64",
	}))
	assert.Equal(t, PasswordLimits{Min: 12}, PasswordLimitsForAsk(map[string]string{"minLength": "12"}))
	assert.Equal(t, PasswordLimits{Max: 32}, PasswordLimitsForAsk(map[string]string{"maxLength": "32"}))
	assert.Equal(t, PasswordLimits{}, PasswordLimitsForAsk(map[string]string{"minLength": "0", "maxLength": "abc"}))
}

func TestPasswordLimitsForStore(t *testing.T) {
	assert.Equal(t, PasswordLimits{Min: DefaultMinPasswordLength, Max: DefaultMaxPasswordLength}, PasswordLimitsForStore(nil))
	assert.Equal(t, PasswordLimits{Min: DefaultMinPasswordLength, Max: DefaultMaxPasswordLength}, PasswordLimitsForStore(map[string]string{}))
	assert.Equal(t, PasswordLimits{Min: 12, Max: 32}, PasswordLimitsForStore(map[string]string{
		"minLength": "12",
		"maxLength": "32",
	}))
	assert.Equal(t, PasswordLimits{Min: DefaultMinPasswordLength, Max: DefaultMaxPasswordLength}, PasswordLimitsForStore(map[string]string{
		"minLength": "abc",
		"maxLength": "abc",
	}))
}

func TestValidateNewPassword(t *testing.T) {
	assert.EqualError(t, ValidateNewPassword("", PasswordLimits{}), "Password is required")
	assert.EqualError(t, ValidateNewPassword("", PasswordLimits{Min: 8}), "Password is required")
	assert.NoError(t, ValidateNewPassword("short", PasswordLimits{}))
	assert.EqualError(t, ValidateNewPassword("short12", PasswordLimits{Min: 8}), "Password must be at least 8 characters")
	assert.NoError(t, ValidateNewPassword("short123", PasswordLimits{Min: 8}))
	assert.NoError(t, ValidateNewPassword("        ", PasswordLimits{Min: 8})) // eight spaces are allowed

	long := strings.Repeat("a", 65)
	assert.EqualError(t, ValidateNewPassword(long, PasswordLimits{Max: 64}), "Password must be at most 64 characters")
	assert.NoError(t, ValidateNewPassword(strings.Repeat("a", 64), PasswordLimits{Min: 8, Max: 64}))
	assert.NoError(t, ValidateNewPassword(long, PasswordLimits{})) // no max unless configured
}
