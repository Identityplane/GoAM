package postgres_adapter

import (
	"context"
	"fmt"

	"github.com/gianlucafrei/GoAM/internal/db"
	"github.com/gianlucafrei/GoAM/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClientSessionDB struct {
	db *pgxpool.Pool
}

// NewPostgresClientSessionDB creates a new PostgresClientSessionDB instance
func NewPostgresClientSessionDB(db *pgxpool.Pool) (db.ClientSessionDB, error) {
	// Check if the connection works and client_sessions table exists by executing a query
	_, err := db.Exec(context.Background(), `
		SELECT 1 FROM client_sessions LIMIT 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to check if client_sessions table exists: %w", err)
	}

	return &PostgresClientSessionDB{db: db}, nil
}

func (s *PostgresClientSessionDB) CreateClientSession(ctx context.Context, tenant, realm string, session *model.ClientSession) error {
	// Ensure realm and tenant are matching
	if session.Tenant != tenant || session.Realm != realm {
		return fmt.Errorf("tenant and realm do not match")
	}

	query := `
		INSERT INTO client_sessions (
			tenant, realm, client_session_id, client_id, grant_type,
			access_token_hash, refresh_token_hash, auth_code_hash,
			user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := s.db.Exec(ctx, query,
		session.Tenant,
		session.Realm,
		session.ClientSessionID,
		session.ClientID,
		session.GrantType,
		session.AccessTokenHash,
		session.RefreshTokenHash,
		session.AuthCodeHash,
		session.UserID,
		session.Scope,
		session.LoginSessionJson,
		session.CodeChallenge,
		session.CodeChallengeMethod,
		session.Created,
		session.Expire,
	)
	if err != nil {
		return fmt.Errorf("failed to create client session: %w", err)
	}

	return nil
}

func (s *PostgresClientSessionDB) GetClientSessionByID(ctx context.Context, tenant, realm, sessionID string) (*model.ClientSession, error) {
	query := `
		SELECT tenant, realm, client_session_id, client_id, grant_type,
		       access_token_hash, refresh_token_hash, auth_code_hash,
		       user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		FROM client_sessions
		WHERE tenant = $1 AND realm = $2 AND client_session_id = $3
	`

	var session model.ClientSession
	err := s.db.QueryRow(ctx, query, tenant, realm, sessionID).Scan(
		&session.Tenant,
		&session.Realm,
		&session.ClientSessionID,
		&session.ClientID,
		&session.GrantType,
		&session.AccessTokenHash,
		&session.RefreshTokenHash,
		&session.AuthCodeHash,
		&session.UserID,
		&session.Scope,
		&session.LoginSessionJson,
		&session.CodeChallenge,
		&session.CodeChallengeMethod,
		&session.Created,
		&session.Expire,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client session: %w", err)
	}

	return &session, nil
}

func (s *PostgresClientSessionDB) GetClientSessionByAccessToken(ctx context.Context, tenant, realm, accessTokenHash string) (*model.ClientSession, error) {
	query := `
		SELECT tenant, realm, client_session_id, client_id, grant_type,
		       access_token_hash, refresh_token_hash, auth_code_hash,
		       user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		FROM client_sessions
		WHERE tenant = $1 AND realm = $2 AND access_token_hash = $3
	`

	var session model.ClientSession
	err := s.db.QueryRow(ctx, query, tenant, realm, accessTokenHash).Scan(
		&session.Tenant,
		&session.Realm,
		&session.ClientSessionID,
		&session.ClientID,
		&session.GrantType,
		&session.AccessTokenHash,
		&session.RefreshTokenHash,
		&session.AuthCodeHash,
		&session.UserID,
		&session.Scope,
		&session.LoginSessionJson,
		&session.CodeChallenge,
		&session.CodeChallengeMethod,
		&session.Created,
		&session.Expire,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client session by access token: %w", err)
	}

	return &session, nil
}

func (s *PostgresClientSessionDB) GetClientSessionByRefreshToken(ctx context.Context, tenant, realm, refreshTokenHash string) (*model.ClientSession, error) {
	query := `
		SELECT tenant, realm, client_session_id, client_id, grant_type,
		       access_token_hash, refresh_token_hash, auth_code_hash,
		       user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		FROM client_sessions
		WHERE tenant = $1 AND realm = $2 AND refresh_token_hash = $3
	`

	var session model.ClientSession
	err := s.db.QueryRow(ctx, query, tenant, realm, refreshTokenHash).Scan(
		&session.Tenant,
		&session.Realm,
		&session.ClientSessionID,
		&session.ClientID,
		&session.GrantType,
		&session.AccessTokenHash,
		&session.RefreshTokenHash,
		&session.AuthCodeHash,
		&session.UserID,
		&session.Scope,
		&session.LoginSessionJson,
		&session.CodeChallenge,
		&session.CodeChallengeMethod,
		&session.Created,
		&session.Expire,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client session by refresh token: %w", err)
	}

	return &session, nil
}

func (s *PostgresClientSessionDB) GetClientSessionByAuthCode(ctx context.Context, tenant, realm, authCodeHash string) (*model.ClientSession, error) {
	query := `
		SELECT tenant, realm, client_session_id, client_id, grant_type,
		       access_token_hash, refresh_token_hash, auth_code_hash,
		       user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		FROM client_sessions
		WHERE tenant = $1 AND realm = $2 AND auth_code_hash = $3
	`

	var session model.ClientSession
	err := s.db.QueryRow(ctx, query, tenant, realm, authCodeHash).Scan(
		&session.Tenant,
		&session.Realm,
		&session.ClientSessionID,
		&session.ClientID,
		&session.GrantType,
		&session.AccessTokenHash,
		&session.RefreshTokenHash,
		&session.AuthCodeHash,
		&session.UserID,
		&session.Scope,
		&session.LoginSessionJson,
		&session.CodeChallenge,
		&session.CodeChallengeMethod,
		&session.Created,
		&session.Expire,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client session by auth code: %w", err)
	}

	return &session, nil
}

func (s *PostgresClientSessionDB) ListClientSessions(ctx context.Context, tenant, realm, clientID string) ([]model.ClientSession, error) {
	query := `
		SELECT tenant, realm, client_session_id, client_id, grant_type,
		       access_token_hash, refresh_token_hash, auth_code_hash,
		       user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		FROM client_sessions
		WHERE tenant = $1 AND realm = $2 AND client_id = $3
	`

	rows, err := s.db.Query(ctx, query, tenant, realm, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to list client sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.ClientSession
	for rows.Next() {
		var session model.ClientSession
		err := rows.Scan(
			&session.Tenant,
			&session.Realm,
			&session.ClientSessionID,
			&session.ClientID,
			&session.GrantType,
			&session.AccessTokenHash,
			&session.RefreshTokenHash,
			&session.AuthCodeHash,
			&session.UserID,
			&session.Scope,
			&session.LoginSessionJson,
			&session.CodeChallenge,
			&session.CodeChallengeMethod,
			&session.Created,
			&session.Expire,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client session: %w", err)
		}

		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating client sessions: %w", err)
	}

	return sessions, nil
}

func (s *PostgresClientSessionDB) ListUserClientSessions(ctx context.Context, tenant, realm, userID string) ([]model.ClientSession, error) {
	query := `
		SELECT tenant, realm, client_session_id, client_id, grant_type,
		       access_token_hash, refresh_token_hash, auth_code_hash,
		       user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		FROM client_sessions
		WHERE tenant = $1 AND realm = $2 AND user_id = $3
	`

	rows, err := s.db.Query(ctx, query, tenant, realm, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user client sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.ClientSession
	for rows.Next() {
		var session model.ClientSession
		err := rows.Scan(
			&session.Tenant,
			&session.Realm,
			&session.ClientSessionID,
			&session.ClientID,
			&session.GrantType,
			&session.AccessTokenHash,
			&session.RefreshTokenHash,
			&session.AuthCodeHash,
			&session.UserID,
			&session.Scope,
			&session.LoginSessionJson,
			&session.CodeChallenge,
			&session.CodeChallengeMethod,
			&session.Created,
			&session.Expire,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client session: %w", err)
		}

		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating client sessions: %w", err)
	}

	return sessions, nil
}

func (s *PostgresClientSessionDB) UpdateClientSession(ctx context.Context, tenant, realm string, session *model.ClientSession) error {
	// Ensure realm and tenant are matching
	if session.Tenant != tenant || session.Realm != realm {
		return fmt.Errorf("tenant and realm do not match")
	}

	query := `
		UPDATE client_sessions
		SET client_id = $1,
			grant_type = $2,
			access_token_hash = $3,
			refresh_token_hash = $4,
			auth_code_hash = $5,
			user_id = $6,
			scope = $7,
			login_session_state_json = $8,
			code_challenge = $9,
			code_challenge_method = $10,
			created = $11,
			expire = $12
		WHERE tenant = $13 AND realm = $14 AND client_session_id = $15
	`

	_, err := s.db.Exec(ctx, query,
		session.ClientID,
		session.GrantType,
		session.AccessTokenHash,
		session.RefreshTokenHash,
		session.AuthCodeHash,
		session.UserID,
		session.Scope,
		session.LoginSessionJson,
		session.CodeChallenge,
		session.CodeChallengeMethod,
		session.Created,
		session.Expire,
		session.Tenant,
		session.Realm,
		session.ClientSessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to update client session: %w", err)
	}

	return nil
}

func (s *PostgresClientSessionDB) DeleteClientSession(ctx context.Context, tenant, realm, sessionID string) error {
	query := `
		DELETE FROM client_sessions
		WHERE tenant = $1 AND realm = $2 AND client_session_id = $3
	`

	_, err := s.db.Exec(ctx, query, tenant, realm, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete client session: %w", err)
	}

	return nil
}

func (s *PostgresClientSessionDB) DeleteExpiredClientSessions(ctx context.Context, tenant, realm string) error {
	query := `
		DELETE FROM client_sessions
		WHERE tenant = $1 AND realm = $2 AND expire < NOW()
	`

	_, err := s.db.Exec(ctx, query, tenant, realm)
	if err != nil {
		return fmt.Errorf("failed to delete expired client sessions: %w", err)
	}

	return nil
}
