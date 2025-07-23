package sqlite_adapter

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/db"
	"github.com/Identityplane/GoAM/pkg/model"
)

type SQLiteClientSessionDB struct {
	db *sql.DB
}

// NewClientSessionDB creates a new ClientSessionDB instance
func NewClientSessionDB(db *sql.DB) (db.ClientSessionDB, error) {

	// Check if the connection works and flows table exists by executing a query
	_, err := db.Exec(`
		SELECT 1 FROM client_sessions LIMIT 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to check if flows table exists: %w", err)
	}

	return &SQLiteClientSessionDB{db: db}, nil
}

func (s *SQLiteClientSessionDB) CreateClientSession(ctx context.Context, tenant, realm string, session *model.ClientSession) error {

	// Esnure realm and tenant are matching
	if session.Tenant != tenant || session.Realm != realm {
		return fmt.Errorf("tenant and realm do not match")
	}

	query := `
		INSERT INTO client_sessions (
			tenant, realm, client_session_id, client_id, grant_type,
			access_token_hash, refresh_token_hash, auth_code_hash,
			user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
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
		session.Created.Format(time.RFC3339),
		session.Expire.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to create client session: %w", err)
	}

	return nil
}

func (s *SQLiteClientSessionDB) GetClientSessionByID(ctx context.Context, tenant, realm, sessionID string) (*model.ClientSession, error) {
	query := `
		SELECT tenant, realm, client_session_id, client_id, grant_type,
		       access_token_hash, refresh_token_hash, auth_code_hash,
		       user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		FROM client_sessions
		WHERE tenant = ? AND realm = ? AND client_session_id = ?
	`

	var session model.ClientSession
	var createdStr, expireStr string

	err := s.db.QueryRowContext(ctx, query, tenant, realm, sessionID).Scan(
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
		&createdStr,
		&expireStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client session: %w", err)
	}

	// Parse timestamps
	created, _ := time.Parse(time.RFC3339, createdStr)
	expire, _ := time.Parse(time.RFC3339, expireStr)

	// Convert to local time to match PostgreSQL behavior
	session.Created = created.Local()
	session.Expire = expire.Local()

	return &session, nil
}

func (s *SQLiteClientSessionDB) GetClientSessionByAccessToken(ctx context.Context, tenant, realm, accessTokenHash string) (*model.ClientSession, error) {
	query := `
		SELECT tenant, realm, client_session_id, client_id, grant_type,
		       access_token_hash, refresh_token_hash, auth_code_hash,
		       user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		FROM client_sessions
		WHERE tenant = ? AND realm = ? AND access_token_hash = ?
	`

	var session model.ClientSession
	var createdStr, expireStr string

	err := s.db.QueryRowContext(ctx, query, tenant, realm, accessTokenHash).Scan(
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
		&createdStr,
		&expireStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client session by access token: %w", err)
	}

	// Parse timestamps
	created, _ := time.Parse(time.RFC3339, createdStr)
	expire, _ := time.Parse(time.RFC3339, expireStr)

	// Convert to local time to match PostgreSQL behavior
	session.Created = created.Local()
	session.Expire = expire.Local()

	return &session, nil
}

func (s *SQLiteClientSessionDB) GetClientSessionByRefreshToken(ctx context.Context, tenant, realm, refreshTokenHash string) (*model.ClientSession, error) {
	query := `
		SELECT tenant, realm, client_session_id, client_id, grant_type,
		       access_token_hash, refresh_token_hash, auth_code_hash,
		       user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		FROM client_sessions
		WHERE tenant = ? AND realm = ? AND refresh_token_hash = ?
	`

	var session model.ClientSession
	var createdStr, expireStr string

	err := s.db.QueryRowContext(ctx, query, tenant, realm, refreshTokenHash).Scan(
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
		&createdStr,
		&expireStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client session by refresh token: %w", err)
	}

	// Parse timestamps
	created, _ := time.Parse(time.RFC3339, createdStr)
	expire, _ := time.Parse(time.RFC3339, expireStr)

	// Convert to local time to match PostgreSQL behavior
	session.Created = created.Local()
	session.Expire = expire.Local()

	return &session, nil
}

func (s *SQLiteClientSessionDB) GetClientSessionByAuthCode(ctx context.Context, tenant, realm, authCodeHash string) (*model.ClientSession, error) {
	query := `
		SELECT tenant, realm, client_session_id, client_id, grant_type,
		       access_token_hash, refresh_token_hash, auth_code_hash,
		       user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		FROM client_sessions
		WHERE tenant = ? AND realm = ? AND auth_code_hash = ?
	`

	var session model.ClientSession
	var createdStr, expireStr string

	err := s.db.QueryRowContext(ctx, query, tenant, realm, authCodeHash).Scan(
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
		&createdStr,
		&expireStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client session by auth code: %w", err)
	}

	// Parse timestamps
	created, _ := time.Parse(time.RFC3339, createdStr)
	expire, _ := time.Parse(time.RFC3339, expireStr)

	// Convert to local time to match PostgreSQL behavior
	session.Created = created.Local()
	session.Expire = expire.Local()

	return &session, nil
}

func (s *SQLiteClientSessionDB) ListClientSessions(ctx context.Context, tenant, realm, clientID string) ([]model.ClientSession, error) {
	query := `
		SELECT tenant, realm, client_session_id, client_id, grant_type,
		       access_token_hash, refresh_token_hash, auth_code_hash,
		       user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		FROM client_sessions
		WHERE tenant = ? AND realm = ? AND client_id = ?
	`

	rows, err := s.db.QueryContext(ctx, query, tenant, realm, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to list client sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.ClientSession
	for rows.Next() {
		var session model.ClientSession
		var createdStr, expireStr string

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
			&createdStr,
			&expireStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client session: %w", err)
		}

		// Parse timestamps
		created, _ := time.Parse(time.RFC3339, createdStr)
		expire, _ := time.Parse(time.RFC3339, expireStr)

		// Convert to local time to match PostgreSQL behavior
		session.Created = created.Local()
		session.Expire = expire.Local()

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s *SQLiteClientSessionDB) ListUserClientSessions(ctx context.Context, tenant, realm, userID string) ([]model.ClientSession, error) {
	query := `
		SELECT tenant, realm, client_session_id, client_id, grant_type,
		       access_token_hash, refresh_token_hash, auth_code_hash,
		       user_id, scope, login_session_state_json, code_challenge, code_challenge_method, created, expire
		FROM client_sessions
		WHERE tenant = ? AND realm = ? AND user_id = ?
	`

	rows, err := s.db.QueryContext(ctx, query, tenant, realm, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user client sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.ClientSession
	for rows.Next() {
		var session model.ClientSession
		var createdStr, expireStr string

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
			&createdStr,
			&expireStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client session: %w", err)
		}

		// Parse timestamps
		created, _ := time.Parse(time.RFC3339, createdStr)
		expire, _ := time.Parse(time.RFC3339, expireStr)

		// Convert to local time to match PostgreSQL behavior
		session.Created = created.Local()
		session.Expire = expire.Local()

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s *SQLiteClientSessionDB) UpdateClientSession(ctx context.Context, tenant, realm string, session *model.ClientSession) error {

	// Esnure realm and tenant are matching
	if session.Tenant != tenant || session.Realm != realm {
		return fmt.Errorf("tenant and realm do not match")
	}

	query := `
		UPDATE client_sessions
		SET client_id = ?, grant_type = ?,
		    access_token_hash = ?, refresh_token_hash = ?, auth_code_hash = ?,
		    user_id = ?, scope = ?, login_session_state_json = ?, code_challenge = ?, code_challenge_method = ?,
		    created = ?, expire = ?
		WHERE tenant = ? AND realm = ? AND client_session_id = ?
	`

	_, err := s.db.ExecContext(ctx, query,
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
		session.Created.Format(time.RFC3339),
		session.Expire.Format(time.RFC3339),
		session.Tenant,
		session.Realm,
		session.ClientSessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to update client session: %w", err)
	}

	return nil
}

func (s *SQLiteClientSessionDB) DeleteClientSession(ctx context.Context, tenant, realm, sessionID string) error {
	query := `
		DELETE FROM client_sessions
		WHERE tenant = ? AND realm = ? AND client_session_id = ?
	`

	_, err := s.db.ExecContext(ctx, query, tenant, realm, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete client session: %w", err)
	}

	return nil
}

func (s *SQLiteClientSessionDB) DeleteExpiredClientSessions(ctx context.Context, tenant, realm string) error {
	query := `
		DELETE FROM client_sessions
		WHERE tenant = ? AND realm = ? AND expire < ?
	`

	_, err := s.db.ExecContext(ctx, query, tenant, realm, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to delete expired client sessions: %w", err)
	}

	return nil
}
