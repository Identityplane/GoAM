package nodeyubico

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test vector form Yubico Website https://developers.yubico.com/OTP/Specifications/Test_vectors.html
func TestYubicoOfficialTestVector(t *testing.T) {

	apiKey := "mG5be6ZJU1qBGz24yPh/ESM3UdU="
	expectedSignature := "+ja8S3IjbX593/LAgTBixwPNGX4=" // This is the response signature

	// Parameters from the official test vector
	params := map[string]string{
		"id":    "1",
		"otp":   "vvungrrdhvtklknvrtvuvbbkeidikkvgglrvdgrfcdft",
		"nonce": "jrFwbaYFhn0HoxZIsd9LQ6w2ceU",
	}

	httpClient := &YubicoHttpClient{
		apiKey: apiKey,
	}

	// Generate signature
	actualSignature, err := httpClient.generateHmacSignature(params)
	assert.NoError(t, err, "Should be able to generate signature")

	t.Logf("Expected signature: %s", expectedSignature)
	t.Logf("Actual signature:   %s", actualSignature)
	t.Logf("Parameters: %+v", params)

	assert.Equal(t, expectedSignature, actualSignature, "Generated signature should match the official test vector")
}
