package db

import (
	"context"

	"github.com/Identityplane/GoAM/pkg/model"
)

// RealmDB interface for realm database operations
type RealmDB interface {
	// CreateRealm creates a new realm
	CreateRealm(ctx context.Context, realm model.Realm) error

	// GetRealm retrieves a realm by its tenant and realm name
	GetRealm(ctx context.Context, tenant, realm string) (*model.Realm, error)

	// UpdateRealm updates an existing realm
	UpdateRealm(ctx context.Context, realm *model.Realm) error

	// ListRealms lists all realms for a tenant
	ListRealms(ctx context.Context, tenant string) ([]model.Realm, error)

	// ListRealms lists all realms for all tenants
	ListAllRealms(ctx context.Context) ([]model.Realm, error)

	// DeleteRealm deletes a realm if it's empty
	DeleteRealm(ctx context.Context, tenant, realm string) error
}
