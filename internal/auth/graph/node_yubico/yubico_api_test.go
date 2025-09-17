package nodeyubico

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"strings"
	"testing"
)

func TestVerifyYubicoOtp_StatusOK_ValidHmac(t *testing.T) {
	verifier := &YubicoVerifier{
		apiUrl:    "https://api.yubico.com/wsapi/2.0/verify",
		clientId:  "12345",
		apiKey:    "dGVzdC1hcGkta2V5",
		yubicoApi: &mockYubicoAPI{},
	}

	// Test with a valid OTP
	otp := "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil"

	publicId, err := verifier.VerifyYubicoOtp(otp)
	if err != nil {
		t.Fatalf("VerifyYubicoOtp failed: %v", err)
	}

	expectedPublicId := otp[:12]
	if publicId != expectedPublicId {
		t.Errorf("Expected public ID %s, got %s", expectedPublicId, publicId)
	}
}

func TestVerifyYubicoOtp_StatusOK_InvalidHmac(t *testing.T) {
	verifier := &YubicoVerifier{
		apiUrl:    "https://api.yubico.com/wsapi/2.0/verify",
		clientId:  "12345",
		apiKey:    "dGVzdC1hcGkta2V5",
		yubicoApi: &mockYubicoAPIWithInvalidHmac{},
	}

	otp := "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil"

	_, err := verifier.VerifyYubicoOtp(otp)
	if err == nil {
		t.Fatal("Expected HMAC verification to fail, but it succeeded")
	}

	expectedErrorPrefix := "HMAC signature verification failed"
	if !strings.Contains(err.Error(), expectedErrorPrefix) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedErrorPrefix, err.Error())
	}
}

func TestVerifyYubicoOtp_StatusReplayedOtp(t *testing.T) {
	verifier := &YubicoVerifier{
		apiUrl:    "https://api.yubico.com/wsapi/2.0/verify",
		clientId:  "12345",
		apiKey:    "dGVzdC1hcGkta2V5",
		yubicoApi: &mockYubicoAPIWithReplayedOtp{},
	}

	otp := "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil"

	_, err := verifier.VerifyYubicoOtp(otp)
	if err == nil {
		t.Fatal("Expected verification to fail due to replayed OTP, but it succeeded")
	}

	expectedError := "yubikey verification failed: REPLAYED_OTP"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// Mock implementation for testing with valid HMAC
type mockYubicoAPI struct{}

func (m *mockYubicoAPI) Verify(id, otp, nonce string) (YubicoApiResponse, error) {
	// Generate a valid HMAC for testing
	apiKey := "dGVzdC1hcGkta2V5"
	apiKeyBytes, _ := base64.StdEncoding.DecodeString(apiKey)

	// Generate HMAC signature
	mac := hmac.New(sha1.New, apiKeyBytes)
	queryString := "nonce=" + nonce + "&otp=" + otp + "&sl=100&status=OK&t=2025-01-01T12:00:00Z0000"
	mac.Write([]byte(queryString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return YubicoApiResponse{
		Hmac:      signature,
		OTP:       otp,
		Nonce:     nonce,
		Status:    "OK",
		Timestamp: "2025-01-01T12:00:00Z0000",
		Sl:        100,
	}, nil
}

// Mock implementation for testing with invalid HMAC
type mockYubicoAPIWithInvalidHmac struct{}

func (m *mockYubicoAPIWithInvalidHmac) Verify(id, otp, nonce string) (YubicoApiResponse, error) {
	return YubicoApiResponse{
		Hmac:      "invalid-hmac-signature",
		OTP:       otp,
		Nonce:     nonce,
		Status:    "OK",
		Timestamp: "2025-01-01T12:00:00Z0000",
		Sl:        100,
	}, nil
}

// Mock implementation for testing replayed OTP
type mockYubicoAPIWithReplayedOtp struct{}

func (m *mockYubicoAPIWithReplayedOtp) Verify(id, otp, nonce string) (YubicoApiResponse, error) {
	return YubicoApiResponse{
		OTP:    otp,
		Nonce:  nonce,
		Status: "REPLAYED_OTP",
	}, nil
}
