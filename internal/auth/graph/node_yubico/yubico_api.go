package node_yubico

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/Identityplane/GoAM/internal/lib"
)

const (
	DEFAULT_YUBICO_API_URL = "https://api.yubico.com/wsapi/2.0/verify"
)

/*
Example response format:
h=lR22a1+aZ55S/xVvG3S3AsFG2/c=
t=2025-06-05T20:55:10Z0814
otp=vvcijgklnrbfditdennrucunkbthghhbbgelblvifgkn
nonce=aeBieT6Nake2i4N4iim7sheeChii6u4V
sl=100
status=OK
*/
type YubicoApiResponse struct {
	Hmac              string `json:"h"`
	Timestamp         string `json:"t"`
	OTP               string `json:"otp"`
	Nonce             string `json:"nonce"`
	Sl                int    `json:"sl"`
	Status            string `json:"status"`
	TimestampInternal string `json:"timestamp,omitempty"` // YubiKey internal timestamp
	SessionCounter    string `json:"sessioncounter,omitempty"`
	SessionUse        string `json:"sessionuse,omitempty"`
}

// YubicoVerifier implements the Yubikey verification
type YubicoVerifier struct {
	apiUrl    string
	clientId  string
	apiKey    string
	yubicoApi yubicoApiInterface
}

// NewHttpYubicoVerifier creates a new YubicoVerifier instance using a http client
func NewHttpYubicoVerifier(apiUrl, clientId, apiKey string) *YubicoVerifier {
	return NewYubicoVerifier(apiUrl, clientId, apiKey, newYubicoHttpClient(apiUrl, clientId, apiKey))
}

// NewYubicoVerifier creates a new YubicoVerifier instance
func NewYubicoVerifier(apiUrl, clientId, apiKey string, yubicoApi yubicoApiInterface) *YubicoVerifier {
	return &YubicoVerifier{
		apiUrl:    apiUrl,
		clientId:  clientId,
		apiKey:    apiKey,
		yubicoApi: yubicoApi,
	}
}

// getYubikeyVerifier is a function that returns a YubicoVerifier instance
// For testing this can be overwritten with a mock implementation
var getYubikeyVerifier = func(apiUrl, clientId, apiKey string) *YubicoVerifier {
	return NewHttpYubicoVerifier(apiUrl, clientId, apiKey)
}

// VerifyYubicoOtp verifies an OTP and returns the public ID, success status, and any internal error
// Returns: publicId (string), ok (bool), err (error)
// - publicId: The public ID of the YubiKey if verification succeeds
// - ok: true if verification succeeds, false if OTP is invalid/replayed/etc (but no internal error)
// - err: only set for internal errors like HTTP failures, nil otherwise
// See https://developers.yubico.com/OTP/Specifications/OTP_validation_protocol.html
func (v *YubicoVerifier) VerifyYubicoOtp(otp string) (string, bool, error) {

	// use a secure random number generator to generate a 16 byte hex string
	nonce, err := lib.GenerateRandomBytes(16)
	if err != nil {
		return "", false, err
	}

	// convert the nonce to a 32 character hex string
	nonceStr := hex.EncodeToString(nonce)

	response, err := v.yubicoApi.Verify(v.clientId, otp, nonceStr)
	if err != nil {
		return "", false, err
	}

	// If the status is not OK we return false (validation failure, not internal error)
	if response.Status != "OK" {
		return "", false, nil
	}

	// Validate HMAC signature if present
	if response.Hmac != "" {
		if err := v.verifyHmacSignature(response); err != nil {
			return "", false, nil // HMAC verification failure is a validation failure, not internal error
		}
	}

	// Verify that the OTP in the response matches the one we sent
	if response.OTP != otp {
		return "", false, nil // OTP mismatch is a validation failure, not internal error
	}

	// Verify that the nonce in the response matches the one we sent
	if response.Nonce != nonceStr {
		return "", false, nil // Nonce mismatch is a validation failure, not internal error
	}

	// Extract the public ID from the OTP (first 12 characters)
	if len(otp) < 12 {
		return "", false, nil // Invalid OTP format is a validation failure, not internal error
	}

	publicId := otp[:12]
	return publicId, true, nil
}

// generateHmacSignature generates an HMAC-SHA1 signature for the given parameters
func (v *YubicoVerifier) generateHmacSignature(params map[string]string) (string, error) {
	// Decode the base64 API key
	apiKeyBytes, err := base64.StdEncoding.DecodeString(v.apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode API key: %w", err)
	}

	// Sort parameters alphabetically by key
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build the query string
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	queryString := strings.Join(parts, "&")

	// Generate HMAC-SHA1 signature
	mac := hmac.New(sha1.New, apiKeyBytes)
	mac.Write([]byte(queryString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature, nil
}

// verifyHmacSignature verifies the HMAC signature in the response
func (v *YubicoVerifier) verifyHmacSignature(response YubicoApiResponse) error {
	// Create a map of all response parameters except the signature itself
	params := make(map[string]string)

	if response.OTP != "" {
		params["otp"] = response.OTP
	}
	if response.Nonce != "" {
		params["nonce"] = response.Nonce
	}
	if response.Timestamp != "" {
		params["t"] = response.Timestamp
	}
	if response.Status != "" {
		params["status"] = response.Status
	}
	if response.Sl > 0 {
		params["sl"] = fmt.Sprintf("%d", response.Sl)
	}
	if response.TimestampInternal != "" {
		params["timestamp"] = response.TimestampInternal
	}
	if response.SessionCounter != "" {
		params["sessioncounter"] = response.SessionCounter
	}
	if response.SessionUse != "" {
		params["sessionuse"] = response.SessionUse
	}

	// Generate the expected signature
	expectedSignature, err := v.generateHmacSignature(params)
	if err != nil {
		return fmt.Errorf("failed to generate expected signature: %w", err)
	}

	// Compare signatures
	if !hmac.Equal([]byte(response.Hmac), []byte(expectedSignature)) {
		return fmt.Errorf("signature mismatch: expected %s, got %s", expectedSignature, response.Hmac)
	}

	return nil
}

// This interface implements the yubico validation http api.
type yubicoApiInterface interface {
	/*
		id:  Your numeric Client ID obtained from the API key generator.
		otp: The 44-character ModHex encoded One-Time Password generated by a YubiKey.
		nonce: A unique, random string (16-40 alphanumeric characters) generated by your client for this request. It is a critical security measure to prevent replay attacks and ensure response integrity.

		BAD_OTP: The OTP is invalid.
		REPLAYED_OTP: The OTP has already been seen by the validation server.
		BAD_SIGNATURE: The request was signed with an API key, but the signature was invalid.
		MISSING_PARAMETER: One of the required parameters (`id`, `otp`, or `nonce`) was missing from the request.
		NO_SUCH_CLIENT: The Client ID (`id=...`) does not exist.
		BACKEND_ERROR: An unexpected error occurred on the YubiCloud server. Your client should retry later.
	*/
	Verify(id, otp, nonce string) (YubicoApiResponse, error)
}

// yubicoHttpClient implements YubicoApiInterface using HTTP requests
type yubicoHttpClient struct {
	apiUrl   string
	clientId string
	apiKey   string
	client   *http.Client
}

// newYubicoHttpClient creates a new HTTP client for Yubico API
func newYubicoHttpClient(apiUrl, clientId, apiKey string) *yubicoHttpClient {
	return &yubicoHttpClient{
		apiUrl:   apiUrl,
		clientId: clientId,
		apiKey:   apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Verify implements the YubicoApiInterface by making an HTTP request to the Yubico API
func (c *yubicoHttpClient) Verify(id, otp, nonce string) (YubicoApiResponse, error) {
	// Build query parameters
	params := url.Values{}
	params.Set("id", id)
	params.Set("otp", otp)
	params.Set("nonce", nonce)
	params.Set("timestamp", "1") // Request timestamp and session counter information

	// Build the full URL
	requestURL := c.apiUrl + "?" + params.Encode()

	// Make the HTTP request
	resp, err := c.client.Get(requestURL)
	if err != nil {
		return YubicoApiResponse{}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return YubicoApiResponse{}, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return YubicoApiResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the response
	response, err := c.parseResponse(string(body))
	if err != nil {
		return YubicoApiResponse{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return response, nil
}

// generateHmacSignature generates an HMAC-SHA1 signature for the given parameters
func (c *yubicoHttpClient) generateHmacSignature(params map[string]string) (string, error) {
	// Decode the base64 API key
	apiKeyBytes, err := base64.StdEncoding.DecodeString(c.apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode API key: %w", err)
	}

	// Sort parameters alphabetically by key
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build the query string
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	queryString := strings.Join(parts, "&")

	// Generate HMAC-SHA1 signature
	mac := hmac.New(sha1.New, apiKeyBytes)
	mac.Write([]byte(queryString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature, nil
}

// parseResponse parses the Yubico API response format (key=value pairs separated by CRLF)
func (c *yubicoHttpClient) parseResponse(responseBody string) (YubicoApiResponse, error) {
	response := YubicoApiResponse{}

	// Split by lines (CRLF or LF)
	lines := strings.Split(strings.ReplaceAll(responseBody, "\r\n", "\n"), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split by first '=' to handle values that might contain '='
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "h":
			response.Hmac = value
		case "t":
			response.Timestamp = value
		case "otp":
			response.OTP = value
		case "nonce":
			response.Nonce = value
		case "sl":
			fmt.Sscanf(value, "%d", &response.Sl)
		case "status":
			response.Status = value
		case "timestamp":
			response.TimestampInternal = value
		case "sessioncounter":
			response.SessionCounter = value
		case "sessionuse":
			response.SessionUse = value
		}
	}

	return response, nil
}
