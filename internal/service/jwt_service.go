package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/db"
	"github.com/Identityplane/GoAM/internal/lib/jwt_ec256"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/google/uuid"
)

// JWTService defines the business logic for JWT operations
type JWTService interface {
	// LoadPublicKeys returns the JWKS for a given tenant and realm
	LoadPublicKeys(tenant, realm string) (string, error)

	// SignJWT signs a JWT token with the key for the given tenant and realm
	SignJWT(tenant, realm string, claims map[string]interface{}) (string, error)

	// GenerateKey generates a new key for a tenant/realm
	GenerateKey(tenant, realm string) error

	// RotateKey generates a new key and disables the old one
	RotateKey(tenant, realm string) error

	// getActiveSigningKey returns an active signing key for the given tenant and realm
	// This is an internal method that takes a context
	getActiveSigningKey(ctx context.Context, tenant, realm string) (*model.SigningKey, error)
}

// jwtServiceImpl implements JWTService
type jwtServiceImpl struct {
	signingKeyDB db.SigningKeyDB
}

// NewJWTService creates a new JWTService instance
func NewJWTService(signingKeyDB db.SigningKeyDB) JWTService {
	return &jwtServiceImpl{
		signingKeyDB: signingKeyDB,
	}
}

// ensureKeyExists ensures that a key exists for the given tenant and realm
// If no key exists, it generates one
func (s *jwtServiceImpl) ensureKeyExists(ctx context.Context, tenant, realm string) error {
	// Check if we already have an active key
	keys, err := s.signingKeyDB.ListActiveSigningKeys(ctx, tenant, realm)
	if err != nil {
		return fmt.Errorf("failed to list active keys: %w", err)
	}
	if len(keys) > 0 {
		return nil // Active key already exists
	}

	// Generate a new key
	keyID := uuid.New().String()
	privateKey, err := jwt_ec256.GenerateEC256JWK(keyID)
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	// Extract the public key
	publicKey, err := jwt_ec256.ExtractEC256PublicJWK(privateKey)
	if err != nil {
		return fmt.Errorf("failed to extract public key: %w", err)
	}

	// Create the signing key record
	signingKey := model.SigningKey{
		Tenant:             tenant,
		Realm:              realm,
		Kid:                keyID,
		Active:             true,
		Algorithm:          "EC256",
		Implementation:     "plain",
		SigningKeyMaterial: privateKey,
		PublicKeyJWK:       publicKey,
		Created:            time.Now(),
	}

	// Store the key in the database
	if err := s.signingKeyDB.CreateSigningKey(ctx, signingKey); err != nil {
		return fmt.Errorf("failed to store key: %w", err)
	}

	return nil
}

// LoadPublicKeys returns the JWKS for a given tenant and realm
func (s *jwtServiceImpl) LoadPublicKeys(tenant, realm string) (string, error) {
	// Ensure we have a key
	if err := s.ensureKeyExists(context.Background(), tenant, realm); err != nil {
		return "", fmt.Errorf("failed to ensure key exists: %w", err)
	}

	// Get all keys for this tenant/realm, including disabled ones
	keys, err := s.signingKeyDB.ListSigningKeys(context.Background(), tenant, realm)
	if err != nil {
		return "", fmt.Errorf("failed to list keys: %w", err)
	}

	// Create a JWKS set with all keys
	var jwksKeys []interface{}
	for _, key := range keys {
		// Parse the public key JSON
		var keyMap map[string]interface{}
		if err := json.Unmarshal([]byte(key.PublicKeyJWK), &keyMap); err != nil {
			return "", fmt.Errorf("failed to parse public key JSON: %w", err)
		}
		jwksKeys = append(jwksKeys, keyMap)
	}

	// Create the JWKS set
	jwks := map[string]interface{}{
		"keys": jwksKeys,
	}

	// Marshal the JWKS set to JSON
	jwksJSON, err := json.MarshalIndent(jwks, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWKS: %w", err)
	}

	return string(jwksJSON), nil
}

// getActiveSigningKey returns an active signing key for the given tenant and realm
func (s *jwtServiceImpl) getActiveSigningKey(ctx context.Context, tenant, realm string) (*model.SigningKey, error) {
	// Get an active key for this tenant/realm
	keys, err := s.signingKeyDB.ListActiveSigningKeys(ctx, tenant, realm)
	if err != nil {
		return nil, fmt.Errorf("failed to list active keys: %w", err)
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("no active keys found")
	}

	// Use the first active key
	return &keys[0], nil
}

// SignJWT signs a JWT token with the key for the given tenant and realm
func (s *jwtServiceImpl) SignJWT(tenant, realm string, claims map[string]interface{}) (string, error) {
	// Ensure we have a key
	if err := s.ensureKeyExists(context.Background(), tenant, realm); err != nil {
		return "", fmt.Errorf("failed to ensure key exists: %w", err)
	}

	// Get an active signing key
	key, err := s.getActiveSigningKey(context.Background(), tenant, realm)
	if err != nil {
		return "", err
	}

	// Create a signer with the private key
	signer, err := jwt_ec256.NewJWTSignerEC256(key.SigningKeyMaterial)
	if err != nil {
		return "", fmt.Errorf("failed to create signer: %w", err)
	}

	// Sign the token
	token, err := signer.SignEC256(claims)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return token, nil
}

// GenerateKey generates a new key for a tenant/realm
func (s *jwtServiceImpl) GenerateKey(tenant, realm string) error {
	return s.ensureKeyExists(context.Background(), tenant, realm)
}

// RotateKey generates a new key and disables the old one
func (s *jwtServiceImpl) RotateKey(tenant, realm string) error {
	ctx := context.Background()
	// First, disable all existing active keys
	keys, err := s.signingKeyDB.ListActiveSigningKeys(ctx, tenant, realm)
	if err != nil {
		return fmt.Errorf("failed to list active keys: %w", err)
	}

	for _, key := range keys {
		if err := s.signingKeyDB.DisableSigningKey(ctx, tenant, realm, key.Kid); err != nil {
			return fmt.Errorf("failed to disable key %s: %w", key.Kid, err)
		}
	}

	// Then generate a new key
	return s.ensureKeyExists(ctx, tenant, realm)
}
