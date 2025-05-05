package jwt_ec256

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type JWTSignerEC256 struct {
	kid    string
	signer *ecdsa.PrivateKey
	method jwt.SigningMethod
}

func GenerateEC256JWK(keyID string) (string, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate EC key: %w", err)
	}

	jwkKey, err := jwk.FromRaw(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to convert to JWK: %w", err)
	}

	jwkKey.Set(jwk.KeyIDKey, keyID)
	jwkKey.Set(jwk.AlgorithmKey, "ES256")
	jwkKey.Set(jwk.KeyUsageKey, "sig")

	jwkJSON, err := json.MarshalIndent(jwkKey, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWK to JSON: %w", err)
	}

	return string(jwkJSON), nil
}

func ExtractEC256PublicJWK(privateJWKJSON string) (string, error) {

	keys, err := jwk.ParseString(privateJWKJSON)
	if err != nil {
		return "", fmt.Errorf("failed to parse JWKs: %w", err)
	}

	if keys.Len() != 1 {
		return "", fmt.Errorf("expected 1 key, got %d", keys.Len())
	}

	key, ok := keys.Key(0)
	if !ok {
		return "", fmt.Errorf("failed to get key")
	}

	pubKey, err := jwk.PublicKeyOf(key)
	if err != nil {
		return "", fmt.Errorf("failed to extract public key: %w", err)
	}

	pubJSON, err := json.Marshal(pubKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public JWK: %w", err)
	}

	return string(pubJSON), nil
}

func NewJWTSignerEC256(jwkJSON string) (*JWTSignerEC256, error) {

	keys, err := jwk.ParseString(jwkJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWKs: %w", err)
	}

	if keys.Len() != 1 {
		return nil, fmt.Errorf("expected 1 key, got %d", keys.Len())
	}

	key, ok := keys.Key(0)
	if !ok {
		return nil, fmt.Errorf("failed to get key")
	}

	var rawKey interface{}
	if err := key.Raw(&rawKey); err != nil {
		return nil, fmt.Errorf("failed to get raw key: %w", err)
	}

	kid := key.KeyID()
	algStr := key.Algorithm()
	method := jwt.GetSigningMethod(algStr.String())
	if method == nil {
		return nil, fmt.Errorf("unsupported signing method: %s", algStr)
	}

	return &JWTSignerEC256{
		kid:    kid,
		signer: rawKey.(*ecdsa.PrivateKey),
		method: method,
	}, nil
}

func (js *JWTSignerEC256) SignEC256(claims map[string]interface{}) (string, error) {

	token := jwt.New(js.method)
	mapClaims := jwt.MapClaims{}

	for k, v := range claims {
		mapClaims[k] = v
	}

	token.Claims = mapClaims

	if js.kid != "" {
		token.Header["kid"] = js.kid
	}

	return token.SignedString(js.signer)
}
