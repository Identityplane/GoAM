package jwt_ec256

import (
	"encoding/json"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateECJWK(t *testing.T) {
	keyID := "test-key-1"
	jwkJSON, err := GenerateEC256JWK(keyID)
	require.NoError(t, err)
	require.NotEmpty(t, jwkJSON)

	// Verify the JWK structure
	var jwkMap map[string]interface{}
	err = json.Unmarshal([]byte(jwkJSON), &jwkMap)
	require.NoError(t, err)

	assert.Equal(t, keyID, jwkMap["kid"])
	assert.Equal(t, "ES256", jwkMap["alg"])
	assert.Equal(t, "sig", jwkMap["use"])
	assert.NotEmpty(t, jwkMap["d"]) // private key should be present
}

func TestExtractPublicJWKEC256(t *testing.T) {
	// First generate a private JWK
	keyID := "test-key-2"
	privateJWK, err := GenerateEC256JWK(keyID)
	require.NoError(t, err)

	// Extract public JWK
	publicJWK, err := ExtractEC256PublicJWK(privateJWK)
	require.NoError(t, err)
	require.NotEmpty(t, publicJWK)

	// Verify the public JWK structure
	var jwkMap map[string]interface{}
	err = json.Unmarshal([]byte(publicJWK), &jwkMap)
	require.NoError(t, err)

	assert.Equal(t, keyID, jwkMap["kid"])
	assert.Equal(t, "ES256", jwkMap["alg"])
	assert.Equal(t, "sig", jwkMap["use"])
	assert.Empty(t, jwkMap["d"]) // private key should not be present
}

func TestNewEC256JWTSigner(t *testing.T) {
	// Generate a test JWK
	keyID := "test-key-3"
	jwkJSON, err := GenerateEC256JWK(keyID)
	require.NoError(t, err)

	// Create signer
	signer, err := NewJWTSignerEC256(jwkJSON)
	require.NoError(t, err)
	require.NotNil(t, signer)
	assert.Equal(t, keyID, signer.kid)
	assert.NotNil(t, signer.signer)
	assert.NotNil(t, signer.method)
}

func TestEC265JWTSigner_Sign(t *testing.T) {
	// Generate a test JWK
	keyID := "test-key-4"
	jwkJSON, err := GenerateEC256JWK(keyID)
	require.NoError(t, err)

	// Create signer
	signer, err := NewJWTSignerEC256(jwkJSON)
	require.NoError(t, err)

	// Test claims
	claims := map[string]interface{}{
		"sub": "test-user",
		"iss": "test-issuer",
	}

	// Sign the token
	tokenString, err := signer.SignEC256(claims)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	// Parse and verify the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Extract public key from JWK
		publicJWK, err := ExtractEC256PublicJWK(jwkJSON)
		require.NoError(t, err)

		keys, err := jwk.ParseString(publicJWK)
		require.NoError(t, err)

		key, ok := keys.Key(0)
		require.True(t, ok)

		var rawKey interface{}
		err = key.Raw(&rawKey)
		require.NoError(t, err)

		return rawKey, nil
	})

	require.NoError(t, err)
	assert.True(t, token.Valid)

	// Verify claims
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, "test-user", mapClaims["sub"])
	assert.Equal(t, "test-issuer", mapClaims["iss"])
}
