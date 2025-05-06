package db

import (
	"context"
	"goiam/internal/model"
)

// ClientSessionDB defines the database operations for client sessions
type ClientSessionDB interface {
	// CreateClientSession creates a new client session
	CreateClientSession(ctx context.Context, tenant, realm string, session *model.ClientSession) error

	// GetClientSessionByID returns a client session by its ID
	GetClientSessionByID(ctx context.Context, tenant, realm, sessionID string) (*model.ClientSession, error)

	// GetClientSessionByAccessToken returns a client session by access token hash
	GetClientSessionByAccessToken(ctx context.Context, tenant, realm, accessTokenHash string) (*model.ClientSession, error)

	// GetClientSessionByRefreshToken returns a client session by refresh token hash
	GetClientSessionByRefreshToken(ctx context.Context, tenant, realm, refreshTokenHash string) (*model.ClientSession, error)

	// GetClientSessionByAuthCode returns a client session by auth code hash
	GetClientSessionByAuthCode(ctx context.Context, tenant, realm, authCodeHash string) (*model.ClientSession, error)

	// ListClientSessions returns all client sessions for a client
	ListClientSessions(ctx context.Context, tenant, realm, clientID string) ([]model.ClientSession, error)

	// ListUserClientSessions returns all client sessions for a user
	ListUserClientSessions(ctx context.Context, tenant, realm, userID string) ([]model.ClientSession, error)

	// UpdateClientSession updates an existing client session
	UpdateClientSession(ctx context.Context, tenant, realm string, session *model.ClientSession) error

	// DeleteClientSession deletes a client session
	DeleteClientSession(ctx context.Context, tenant, realm, sessionID string) error

	// DeleteExpiredClientSessions deletes all expired client sessions
	DeleteExpiredClientSessions(ctx context.Context, tenant, realm string) error
}
