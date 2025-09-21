package node_yubico

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

func TestVerifyYubicoOtp_StatusOK_ValidHmac(t *testing.T) {
	verifier := &YubicoVerifier{
		apiUrl:    DEFAULT_YUBICO_API_URL,
		clientId:  "12345",
		apiKey:    "dGVzdC1hcGkta2V5",
		yubicoApi: &mockYubicoAPI{},
	}

	// Test with a valid OTP
	otp := "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil"

	publicId, ok, err := verifier.VerifyYubicoOtp(otp)
	if err != nil {
		t.Fatalf("VerifyYubicoOtp failed with internal error: %v", err)
	}
	if !ok {
		t.Fatal("Expected verification to succeed, but it failed")
	}

	expectedPublicId := otp[:12]
	if publicId != expectedPublicId {
		t.Errorf("Expected public ID %s, got %s", expectedPublicId, publicId)
	}
}

func TestVerifyYubicoOtp_StatusOK_InvalidHmac(t *testing.T) {
	verifier := &YubicoVerifier{
		apiUrl:    DEFAULT_YUBICO_API_URL,
		clientId:  "12345",
		apiKey:    "dGVzdC1hcGkta2V5",
		yubicoApi: &mockYubicoAPIWithInvalidHmac{},
	}

	otp := "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil"

	_, ok, err := verifier.VerifyYubicoOtp(otp)
	if err != nil {
		t.Fatalf("Expected no internal error, got: %v", err)
	}
	if ok {
		t.Fatal("Expected HMAC verification to fail, but it succeeded")
	}
}

func TestVerifyYubicoOtp_StatusReplayedOtp(t *testing.T) {
	verifier := &YubicoVerifier{
		apiUrl:    DEFAULT_YUBICO_API_URL,
		clientId:  "12345",
		apiKey:    "dGVzdC1hcGkta2V5",
		yubicoApi: &mockYubicoAPIWithReplayedOtp{},
	}

	otp := "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil"

	_, ok, err := verifier.VerifyYubicoOtp(otp)
	if err != nil {
		t.Fatalf("Expected no internal error, got: %v", err)
	}
	if ok {
		t.Fatal("Expected verification to fail due to replayed OTP, but it succeeded")
	}
}

func TestVerifyYubicoOtp_InternalError(t *testing.T) {
	verifier := &YubicoVerifier{
		apiUrl:    DEFAULT_YUBICO_API_URL,
		clientId:  "12345",
		apiKey:    "dGVzdC1hcGkta2V5",
		yubicoApi: &mockYubicoAPIWithInternalError{},
	}

	otp := "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil"

	_, ok, err := verifier.VerifyYubicoOtp(otp)
	if err == nil {
		t.Fatal("Expected internal error, but got none")
	}
	if ok {
		t.Fatal("Expected verification to fail due to internal error")
	}

	expectedError := "HTTP request failed"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestVerifyYubicoOtp_InvalidOtpFormat(t *testing.T) {
	verifier := &YubicoVerifier{
		apiUrl:    DEFAULT_YUBICO_API_URL,
		clientId:  "12345",
		apiKey:    "dGVzdC1hcGkta2V5",
		yubicoApi: &mockYubicoAPI{},
	}

	// Test with an OTP that's too short
	otp := "short"

	_, ok, err := verifier.VerifyYubicoOtp(otp)
	if err != nil {
		t.Fatalf("Expected no internal error, got: %v", err)
	}
	if ok {
		t.Fatal("Expected verification to fail due to invalid OTP format")
	}
}

func TestVerifyYubicoOtp_Otpmismatch(t *testing.T) {
	verifier := &YubicoVerifier{
		apiUrl:    DEFAULT_YUBICO_API_URL,
		clientId:  "12345",
		apiKey:    "dGVzdC1hcGkta2V5",
		yubicoApi: &mockYubicoAPIWithOtpMismatch{},
	}

	otp := "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil"

	_, ok, err := verifier.VerifyYubicoOtp(otp)
	if err != nil {
		t.Fatalf("Expected no internal error, got: %v", err)
	}
	if ok {
		t.Fatal("Expected verification to fail due to OTP mismatch")
	}
}

func TestVerifyYubicoOtp_NonceMismatch(t *testing.T) {
	verifier := &YubicoVerifier{
		apiUrl:    DEFAULT_YUBICO_API_URL,
		clientId:  "12345",
		apiKey:    "dGVzdC1hcGkta2V5",
		yubicoApi: &mockYubicoAPIWithNonceMismatch{},
	}

	otp := "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil"

	_, ok, err := verifier.VerifyYubicoOtp(otp)
	if err != nil {
		t.Fatalf("Expected no internal error, got: %v", err)
	}
	if ok {
		t.Fatal("Expected verification to fail due to nonce mismatch")
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

// Mock implementation for testing internal errors
type mockYubicoAPIWithInternalError struct{}

func (m *mockYubicoAPIWithInternalError) Verify(id, otp, nonce string) (YubicoApiResponse, error) {
	return YubicoApiResponse{}, errors.New("HTTP request failed")
}

// Mock implementation for testing OTP mismatch
type mockYubicoAPIWithOtpMismatch struct{}

func (m *mockYubicoAPIWithOtpMismatch) Verify(id, otp, nonce string) (YubicoApiResponse, error) {
	return YubicoApiResponse{
		OTP:    "different-otp",
		Nonce:  nonce,
		Status: "OK",
	}, nil
}

// Mock implementation for testing nonce mismatch
type mockYubicoAPIWithNonceMismatch struct{}

func (m *mockYubicoAPIWithNonceMismatch) Verify(id, otp, nonce string) (YubicoApiResponse, error) {
	return YubicoApiResponse{
		OTP:    otp,
		Nonce:  "different-nonce",
		Status: "OK",
	}, nil
}
