package service

import (
	"fmt"
	"goiam/internal/lib/jwt_ec256"
	"sync"
)

// JWTService handles JWT token signing and JWKS operations
type JWTService struct {
	mu sync.RWMutex
	// In-memory storage of keys per tenant/realm
	keys map[string]map[string]string // tenant:realm -> private key JWK
}

// NewJWTService creates a new JWTService instance
func NewJWTService() *JWTService {
	return &JWTService{
		keys: make(map[string]map[string]string),
	}
}

// ensureKeyExists ensures that a key exists for the given tenant and realm
// If no key exists, it generates one
func (s *JWTService) ensureKeyExists(tenant, realm string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if we already have a key
	if realmKeys, ok := s.keys[tenant]; ok {
		if _, ok := realmKeys[realm]; ok {
			return nil // Key already exists
		}
	}

	// Generate a new key
	keyID := fmt.Sprintf("%s:%s", tenant, realm)
	privateKey, err := jwt_ec256.GenerateEC256JWK(keyID)
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	// Store the key
	if _, ok := s.keys[tenant]; !ok {
		s.keys[tenant] = make(map[string]string)
	}
	s.keys[tenant][realm] = privateKey

	return nil
}

// LoadPublicKeys returns the JWKS for a given tenant and realm
func (s *JWTService) LoadPublicKeys(tenant, realm string) (string, error) {
	// Ensure we have a key
	if err := s.ensureKeyExists(tenant, realm); err != nil {
		return "", fmt.Errorf("failed to ensure key exists: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get the private key for this tenant/realm
	realmKeys := s.keys[tenant]
	privateKey := realmKeys[realm]

	// Extract the public key from the private key
	publicKey, err := jwt_ec256.ExtractEC256PublicJWK(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to extract public key: %w", err)
	}

	return publicKey, nil
}

// SignJWT signs a JWT token with the key for the given tenant and realm
func (s *JWTService) SignJWT(tenant, realm string, claims map[string]interface{}) (string, error) {
	// Ensure we have a key
	if err := s.ensureKeyExists(tenant, realm); err != nil {
		return "", fmt.Errorf("failed to ensure key exists: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get the private key for this tenant/realm
	realmKeys := s.keys[tenant]
	privateKey := realmKeys[realm]

	// Create a signer with the private key
	signer, err := jwt_ec256.NewJWTSignerEC256(privateKey)
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
// This is now just a convenience method that calls ensureKeyExists
func (s *JWTService) GenerateKey(tenant, realm string) error {
	return s.ensureKeyExists(tenant, realm)
}
