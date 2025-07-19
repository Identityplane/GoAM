package db

import (
	"context"

	"github.com/Identityplane/GoAM/internal/model"
)

// ApplicationDB interface for application database operations
type ApplicationDB interface {
	// CreateApplication creates a new application
	CreateApplication(ctx context.Context, app model.Application) error

	// GetApplication retrieves an application by its tenant, realm and id
	GetApplication(ctx context.Context, tenant, realm, id string) (*model.Application, error)

	// UpdateApplication updates an existing application
	UpdateApplication(ctx context.Context, app *model.Application) error

	// ListApplications lists all applications for a tenant and realm
	ListApplications(ctx context.Context, tenant, realm string) ([]model.Application, error)

	// ListAllApplications lists all applications across all tenants and realms
	ListAllApplications(ctx context.Context) ([]model.Application, error)

	// DeleteApplication deletes an application
	DeleteApplication(ctx context.Context, tenant, realm, id string) error
}
