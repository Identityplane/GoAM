package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "mySecurePassword123!",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "Long password",
			password: "thisIsAVeryLongPasswordThatShouldAlsoWorkCorrectly123!@#$%^&*()",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Hash the password
			hashed, err := HashPassword(tc.password)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			// Verify no error occurred
			assert.NoError(t, err)
			assert.NotEmpty(t, hashed)
			assert.NotEqual(t, tc.password, hashed) // Ensure the hash is different from the original password

			// Verify the hash can be compared with the original password
			err = ComparePassword(tc.password, hashed)
			assert.NoError(t, err)

			// Verify the hash cannot be compared with a different password
			err = ComparePassword("wrongPassword", hashed)
			assert.Error(t, err)
		})
	}
}

func TestComparePassword(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		password       string
		hashedPassword string
		wantErr        bool
	}{
		{
			name:           "Matching password",
			password:       "testPassword123",
			hashedPassword: "$2a$10$abcdefghijklmnopqrstuvwxyz1234567890", // This is an invalid hash format
			wantErr:        true,                                          // Should error because the hash is invalid
		},
		{
			name:           "Empty password",
			password:       "",
			hashedPassword: "$2a$10$abcdefghijklmnopqrstuvwxyz1234567890",
			wantErr:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// First create a valid hash for the password
			hashed, err := HashPassword(tc.password)
			assert.NoError(t, err)

			// Test with the correct password
			err = ComparePassword(tc.password, hashed)
			assert.NoError(t, err)

			// Test with an incorrect password
			err = ComparePassword("wrongPassword", hashed)
			assert.Error(t, err)

			// Test with an invalid hash
			err = ComparePassword(tc.password, "invalidHash")
			assert.Error(t, err)
		})
	}
}
