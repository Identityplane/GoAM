package db

import (
	"context"

	"github.com/gianlucafrei/GoAM/internal/model"
)

// SigningKeyDB interface for signing key database operations
type SigningKeyDB interface {
	// CreateSigningKey creates a new signing key
	CreateSigningKey(ctx context.Context, key model.SigningKey) error

	// GetSigningKey retrieves a signing key by its tenant, realm and kid
	GetSigningKey(ctx context.Context, tenant, realm, kid string) (*model.SigningKey, error)

	// UpdateSigningKey updates an existing signing key
	UpdateSigningKey(ctx context.Context, key *model.SigningKey) error

	// ListSigningKeys lists all signing keys for a tenant and realm
	ListSigningKeys(ctx context.Context, tenant, realm string) ([]model.SigningKey, error)

	// ListActiveSigningKeys lists all active signing keys for a tenant and realm
	ListActiveSigningKeys(ctx context.Context, tenant, realm string) ([]model.SigningKey, error)

	// DisableSigningKey disables a signing key
	DisableSigningKey(ctx context.Context, tenant, realm, kid string) error

	// DeleteSigningKey deletes a signing key
	DeleteSigningKey(ctx context.Context, tenant, realm, kid string) error
}
