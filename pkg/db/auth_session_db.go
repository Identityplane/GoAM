package db

import (
	"context"

	"github.com/Identityplane/GoAM/pkg/model"
)

// AuthSessionDB defines the interface for authentication session storage
type AuthSessionDB interface {
	// CreateOrUpdateAuthSession creates a new authentication session or updates an existing one
	CreateOrUpdateAuthSession(ctx context.Context, session *model.PersistentAuthSession) error

	// GetAuthSessionByID retrieves an authentication session by its ID
	GetAuthSessionByID(ctx context.Context, tenant, realm, runID string) (*model.PersistentAuthSession, error)

	// GetAuthSessionByHash retrieves an authentication session by its hash
	GetAuthSessionByHash(ctx context.Context, tenant, realm, sessionIDHash string) (*model.PersistentAuthSession, error)

	// ListAuthSessions lists all authentication sessions for a specific tenant and realm
	ListAuthSessions(ctx context.Context, tenant, realm string) ([]model.PersistentAuthSession, error)

	// ListAllAuthSessions lists all authentication sessions for a specific tenant
	ListAllAuthSessions(ctx context.Context, tenant string) ([]model.PersistentAuthSession, error)

	// DeleteAuthSession deletes an authentication session
	DeleteAuthSession(ctx context.Context, tenant, realm, sessionIDHash string) error

	// DeleteExpiredAuthSessions deletes all expired authentication sessions
	DeleteExpiredAuthSessions(ctx context.Context, tenant, realm string) error
}
