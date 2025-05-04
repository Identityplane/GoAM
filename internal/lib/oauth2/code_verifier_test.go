package oauth2

import (
	"testing"
)

func TestCodeVerifier(t *testing.T) {
	// Test case from the example
	codeVerifier := "2603656fe5db4ebbb8b463f6b2ccc5e6fed1dbb78127467783319c69a50c5a3576a035dd350344e2bbc249415d138522"
	expectedChallenge := "7-e4UfUmqes-3CO73pDP9XRuyYBU_HuCWPxdhZm2k7Q"

	// Test validation
	if err := ValidateCodeVerifier(codeVerifier); err != nil {
		t.Errorf("ValidateCodeVerifier failed: %v", err)
	}

	// Test challenge generation
	challenge, err := GenerateCodeChallenge(codeVerifier)
	if err != nil {
		t.Errorf("GenerateCodeChallenge failed: %v", err)
	}

	if challenge != expectedChallenge {
		t.Errorf("Generated challenge does not match expected challenge.\nGot: %s\nWant: %s", challenge, expectedChallenge)
	}

	// Test verification
	valid, err := VerifyCodeChallenge(codeVerifier, expectedChallenge)
	if err != nil {
		t.Errorf("VerifyCodeChallenge failed: %v", err)
	}

	if !valid {
		t.Error("Code verifier and challenge should match")
	}

	// Test invalid verification
	valid, err = VerifyCodeChallenge(codeVerifier, "invalid_challenge")
	if err != nil {
		t.Errorf("VerifyCodeChallenge failed: %v", err)
	}

	if valid {
		t.Error("Code verifier and invalid challenge should not match")
	}
}

func TestValidateCodeVerifier(t *testing.T) {
	tests := []struct {
		name          string
		codeVerifier  string
		expectError   bool
		errorContains string
	}{
		{
			name:         "Valid code verifier",
			codeVerifier: "2603656fe5db4ebbb8b463f6b2ccc5e6fed1dbb78127467783319c69a50c5a3576a035dd350344e2bbc249415d138522",
		},
		{
			name:          "Too short",
			codeVerifier:  "too_short",
			expectError:   true,
			errorContains: "code verifier length must be between 43 and 128 characters",
		},
		{
			name:          "Invalid characters",
			codeVerifier:  "invalid!@#$%^&*()uzöhjvvcuzihjvculzigkjvhluiögkjbhl",
			expectError:   true,
			errorContains: "code verifier must only contain characters in the range [A-Za-z0-9-._~]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCodeVerifier(tt.codeVerifier)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorContains != "" && err.Error() != tt.errorContains {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
