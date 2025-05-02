package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"goiam/internal/db"
	"goiam/internal/model"

	"github.com/google/uuid"
)

// ApplicationService defines the business logic for application operations
type ApplicationService interface {
	// GetApplication returns an application by its ID
	GetApplication(tenant, realm, clientId string) (*model.Application, bool)

	// ListApplications returns all applications for a tenant and realm
	ListApplications(tenant, realm string) ([]model.Application, error)

	// ListAllApplications returns all applications for all realms
	ListAllApplications() ([]model.Application, error)

	// CreateApplication creates a new application
	CreateApplication(tenant, realm string, app model.Application) error

	// UpdateApplication updates an existing application
	UpdateApplication(tenant, realm string, app model.Application) error

	// DeleteApplication deletes an application by its ID
	DeleteApplication(tenant, realm, clientId string) error

	// RegenerateClientSecret generates a new client secret for an application
	RegenerateClientSecret(tenant, realm, clientId string) (string, error)

	// VerifyClientSecret verifies if a client secret matches the stored hash
	VerifyClientSecret(tenant, realm, clientId, clientSecret string) (bool, error)
}

// applicationServiceImpl implements ApplicationService
type applicationServiceImpl struct {
	appsDb db.ApplicationDB
}

// NewApplicationService creates a new ApplicationService instance
func NewApplicationService(appsDb db.ApplicationDB) ApplicationService {
	return &applicationServiceImpl{
		appsDb: appsDb,
	}
}

func (s *applicationServiceImpl) GetApplication(tenant, realm, clientId string) (*model.Application, bool) {
	app, err := s.appsDb.GetApplication(context.Background(), tenant, realm, clientId)
	if err != nil || app == nil {
		return nil, false
	}
	return app, true
}

func (s *applicationServiceImpl) ListApplications(tenant, realm string) ([]model.Application, error) {
	apps, err := s.appsDb.ListApplications(context.Background(), tenant, realm)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}
	return apps, nil
}

func (s *applicationServiceImpl) ListAllApplications() ([]model.Application, error) {
	apps, err := s.appsDb.ListAllApplications(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list all applications: %w", err)
	}
	return apps, nil
}

func (s *applicationServiceImpl) CreateApplication(tenant, realm string, app model.Application) error {
	// Check that client_id is not empty
	if app.ClientId == "" {
		return fmt.Errorf("client_id is empty")
	}

	// Ensure realm and tenant are set correctly
	app.Realm = realm
	app.Tenant = tenant

	// Check if the application already exists
	_, exists := s.GetApplication(tenant, realm, app.ClientId)
	if exists {
		return fmt.Errorf("application with client_id %s already exists", app.ClientId)
	}

	// Create the application in the database
	return s.appsDb.CreateApplication(context.Background(), app)
}

func (s *applicationServiceImpl) UpdateApplication(tenant, realm string, app model.Application) error {
	// Check that client_id is not empty
	if app.ClientId == "" {
		return fmt.Errorf("client_id is empty")
	}

	// Ensure realm and tenant are set correctly
	app.Realm = realm
	app.Tenant = tenant

	// Check if the application exists
	existingApp, exists := s.GetApplication(tenant, realm, app.ClientId)
	if !exists {
		return fmt.Errorf("application with client_id %s not found", app.ClientId)
	}

	// Preserve the original client secret - it cannot be changed through update
	app.ClientSecret = existingApp.ClientSecret

	// Update the application in the database
	return s.appsDb.UpdateApplication(context.Background(), &app)
}

func (s *applicationServiceImpl) DeleteApplication(tenant, realm, clientId string) error {
	// Get the application first to check if it exists
	_, exists := s.GetApplication(tenant, realm, clientId)
	if !exists {
		return fmt.Errorf("application with client_id %s not found", clientId)
	}

	// Delete the application from the database
	return s.appsDb.DeleteApplication(context.Background(), tenant, realm, clientId)
}

func (s *applicationServiceImpl) RegenerateClientSecret(tenant, realm, clientId string) (string, error) {
	// Get the application first to check if it exists
	app, exists := s.GetApplication(tenant, realm, clientId)
	if !exists {
		return "", fmt.Errorf("application with client_id %s not found", clientId)
	}

	// Generate a new UUID v4 for the client secret
	clientSecret := uuid.New().String()
	app.ClientSecret = hashClientSecret(clientSecret)

	// Update the application in the database
	err := s.appsDb.UpdateApplication(context.Background(), app)
	if err != nil {
		return "", fmt.Errorf("failed to update application: %w", err)
	}

	// Return the unhashed client secret
	return clientSecret, nil
}

func (s *applicationServiceImpl) VerifyClientSecret(tenant, realm, clientId, clientSecret string) (bool, error) {
	// Get the application first to check if it exists
	app, exists := s.GetApplication(tenant, realm, clientId)
	if !exists {
		return false, fmt.Errorf("application with client_id %s not found", clientId)
	}

	// If no client secret is set, verification fails
	if app.ClientSecret == "" {
		return false, nil
	}

	// Hash the provided client secret and compare with stored hash
	hashedSecret := hashClientSecret(clientSecret)
	return hashedSecret == app.ClientSecret, nil
}

// hashClientSecret hashes a client secret using SHA-256
func hashClientSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}
