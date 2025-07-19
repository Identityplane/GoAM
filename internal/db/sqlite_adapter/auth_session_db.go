package sqlite_adapter

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/db"
	"github.com/Identityplane/GoAM/internal/model"
)

type SQLiteAuthSessionDB struct {
	db *sql.DB
}

// NewAuthSessionDB creates a new AuthSessionDB instance
func NewAuthSessionDB(db *sql.DB) (db.AuthSessionDB, error) {
	// Check if the connection works and auth_sessions table exists
	_, err := db.Exec(`
		SELECT 1 FROM auth_sessions LIMIT 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to check if auth_sessions table exists: %w", err)
	}

	return &SQLiteAuthSessionDB{db: db}, nil
}

func (s *SQLiteAuthSessionDB) CreateOrUpdateAuthSession(ctx context.Context, session *model.PersistentAuthSession) error {
	// First check if the session exists
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM auth_sessions 
			WHERE tenant = ? AND realm = ? AND session_id_hash = ?
		)`,
		session.Tenant, session.Realm, session.SessionIDHash,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if session exists: %w", err)
	}

	if exists {
		// Update existing session
		query := `
			UPDATE auth_sessions SET
				run_id = ?,
				created_at = ?,
				expires_at = ?,
				session_information = ?
			WHERE tenant = ? AND realm = ? AND session_id_hash = ?
		`
		_, err = s.db.ExecContext(ctx, query,
			session.RunID,
			session.CreatedAt.Format(time.RFC3339),
			session.ExpiresAt.Format(time.RFC3339),
			session.SessionInformation,
			session.Tenant,
			session.Realm,
			session.SessionIDHash,
		)
		if err != nil {
			return fmt.Errorf("failed to update auth session: %w", err)
		}
	} else {
		// Create new session
		query := `
			INSERT INTO auth_sessions (
				tenant, realm, run_id, session_id_hash,
				created_at, expires_at, session_information
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`
		_, err = s.db.ExecContext(ctx, query,
			session.Tenant,
			session.Realm,
			session.RunID,
			session.SessionIDHash,
			session.CreatedAt.Format(time.RFC3339),
			session.ExpiresAt.Format(time.RFC3339),
			session.SessionInformation,
		)
		if err != nil {
			return fmt.Errorf("failed to create auth session: %w", err)
		}
	}

	return nil
}

func (s *SQLiteAuthSessionDB) GetAuthSessionByID(ctx context.Context, tenant, realm, runID string) (*model.PersistentAuthSession, error) {
	query := `
		SELECT tenant, realm, run_id, session_id_hash,
		       created_at, expires_at, session_information
		FROM auth_sessions
		WHERE tenant = ? AND realm = ? AND run_id = ?
	`

	var session model.PersistentAuthSession
	var createdAtStr, expiresAtStr string

	err := s.db.QueryRowContext(ctx, query, tenant, realm, runID).Scan(
		&session.Tenant,
		&session.Realm,
		&session.RunID,
		&session.SessionIDHash,
		&createdAtStr,
		&expiresAtStr,
		&session.SessionInformation,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get auth session: %w", err)
	}

	// Parse timestamps
	session.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	session.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAtStr)

	return &session, nil
}

func (s *SQLiteAuthSessionDB) GetAuthSessionByHash(ctx context.Context, tenant, realm, sessionIDHash string) (*model.PersistentAuthSession, error) {
	query := `
		SELECT tenant, realm, run_id, session_id_hash,
		       created_at, expires_at, session_information
		FROM auth_sessions
		WHERE tenant = ? AND realm = ? AND session_id_hash = ?
	`

	var session model.PersistentAuthSession
	var createdAtStr, expiresAtStr string

	err := s.db.QueryRowContext(ctx, query, tenant, realm, sessionIDHash).Scan(
		&session.Tenant,
		&session.Realm,
		&session.RunID,
		&session.SessionIDHash,
		&createdAtStr,
		&expiresAtStr,
		&session.SessionInformation,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get auth session by hash: %w", err)
	}

	// Parse timestamps
	session.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	session.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAtStr)

	return &session, nil
}

func (s *SQLiteAuthSessionDB) ListAuthSessions(ctx context.Context, tenant, realm string) ([]model.PersistentAuthSession, error) {
	query := `
		SELECT tenant, realm, run_id, session_id_hash,
		       created_at, expires_at, session_information
		FROM auth_sessions
		WHERE tenant = ? AND realm = ?
	`

	rows, err := s.db.QueryContext(ctx, query, tenant, realm)
	if err != nil {
		return nil, fmt.Errorf("failed to list auth sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.PersistentAuthSession
	for rows.Next() {
		var session model.PersistentAuthSession
		var createdAtStr, expiresAtStr string

		err := rows.Scan(
			&session.Tenant,
			&session.Realm,
			&session.RunID,
			&session.SessionIDHash,
			&createdAtStr,
			&expiresAtStr,
			&session.SessionInformation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan auth session: %w", err)
		}

		// Parse timestamps
		session.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		session.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAtStr)

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s *SQLiteAuthSessionDB) ListAllAuthSessions(ctx context.Context, tenant string) ([]model.PersistentAuthSession, error) {
	query := `
		SELECT tenant, realm, run_id, session_id_hash,
		       created_at, expires_at, session_information
		FROM auth_sessions
		WHERE tenant = ?
	`

	rows, err := s.db.QueryContext(ctx, query, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to list all auth sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.PersistentAuthSession
	for rows.Next() {
		var session model.PersistentAuthSession
		var createdAtStr, expiresAtStr string

		err := rows.Scan(
			&session.Tenant,
			&session.Realm,
			&session.RunID,
			&session.SessionIDHash,
			&createdAtStr,
			&expiresAtStr,
			&session.SessionInformation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan auth session: %w", err)
		}

		// Parse timestamps
		session.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		session.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAtStr)

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s *SQLiteAuthSessionDB) DeleteAuthSession(ctx context.Context, tenant, realm, sessionIDHash string) error {
	query := `
		DELETE FROM auth_sessions
		WHERE tenant = ? AND realm = ? AND session_id_hash = ?
	`

	_, err := s.db.ExecContext(ctx, query, tenant, realm, sessionIDHash)
	if err != nil {
		return fmt.Errorf("failed to delete auth session: %w", err)
	}

	return nil
}

func (s *SQLiteAuthSessionDB) DeleteExpiredAuthSessions(ctx context.Context, tenant, realm string) error {
	query := `
		DELETE FROM auth_sessions
		WHERE tenant = ? AND realm = ? AND expires_at < ?
	`

	_, err := s.db.ExecContext(ctx, query, tenant, realm, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to delete expired auth sessions: %w", err)
	}

	return nil
}
