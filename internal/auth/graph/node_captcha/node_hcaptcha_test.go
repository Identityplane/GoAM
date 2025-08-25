package node_captcha

import (
	"testing"

	"github.com/Identityplane/GoAM/pkg/model"
)

// MockHCaptchaVerifier implements HCaptchaVerifier for testing
type MockHCaptchaVerifier struct {
	shouldVerify bool
}

func (m *MockHCaptchaVerifier) Verify(response, sitekey, secret string) bool {
	return m.shouldVerify
}

func TestRunHcaptchaNode(t *testing.T) {
	// Create mock services
	services := &model.Repositories{}

	// Create test cases
	tests := []struct {
		name           string
		verifier       HCaptchaVerifier
		nodeConfig     map[string]string
		input          map[string]string
		expectedResult string
		expectError    bool
	}{
		{
			name:     "Missing sitekey",
			verifier: &MockHCaptchaVerifier{shouldVerify: true},
			nodeConfig: map[string]string{
				"hcaptcha_secret": "test-secret",
			},
			input: map[string]string{
				"hcaptcha": "test-response",
			},
			expectedResult: "",
			expectError:    true,
		},
		{
			name:     "Missing secret",
			verifier: &MockHCaptchaVerifier{shouldVerify: true},
			nodeConfig: map[string]string{
				"hcaptcha_sitekey": "test-sitekey",
			},
			input: map[string]string{
				"hcaptcha": "test-response",
			},
			expectedResult: "",
			expectError:    true,
		},
		{
			name:     "Empty response",
			verifier: &MockHCaptchaVerifier{shouldVerify: true},
			nodeConfig: map[string]string{
				"hcaptcha_sitekey": "test-sitekey",
				"hcaptcha_secret":  "test-secret",
			},
			input:          map[string]string{},
			expectedResult: "",
			expectError:    false,
		},
		{
			name:     "Successful verification",
			verifier: &MockHCaptchaVerifier{shouldVerify: true},
			nodeConfig: map[string]string{
				"hcaptcha_sitekey": "test-sitekey",
				"hcaptcha_secret":  "test-secret",
			},
			input: map[string]string{
				"hcaptcha": "test-response",
			},
			expectedResult: "success",
			expectError:    false,
		},
		{
			name:     "Failed verification",
			verifier: &MockHCaptchaVerifier{shouldVerify: false},
			nodeConfig: map[string]string{
				"hcaptcha_sitekey": "test-sitekey",
				"hcaptcha_secret":  "test-secret",
			},
			input: map[string]string{
				"hcaptcha": "test-response",
			},
			expectedResult: "failure",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the mock verifier
			SetHCaptchaVerifier(tt.verifier)

			// Create test node
			node := &model.GraphNode{
				CustomConfig: tt.nodeConfig,
			}

			// Create test session
			session := &model.AuthenticationSession{
				Context: make(map[string]string),
			}

			// Run the node
			result, err := RunHcaptchaNode(session, node, tt.input, services)

			// Check error
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check result
			if result.Condition != tt.expectedResult {
				t.Errorf("expected result %q but got %q", tt.expectedResult, result.Condition)
			}

			// Check prompts for empty response case
			if tt.input == nil && result.Prompts == nil {
				t.Error("expected prompts for empty response but got none")
			}
		})
	}
}
