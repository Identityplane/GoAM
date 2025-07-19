package postgres_adapter

import (
	"context"
	"fmt"

	"github.com/Identityplane/GoAM/internal/db"
	"github.com/Identityplane/GoAM/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAuthSessionDB struct {
	db *pgxpool.Pool
}

// NewPostgresAuthSessionDB creates a new PostgresAuthSessionDB instance
func NewPostgresAuthSessionDB(db *pgxpool.Pool) (db.AuthSessionDB, error) {
	// Check if the connection works and auth_sessions table exists
	_, err := db.Exec(context.Background(), `
		SELECT 1 FROM auth_sessions LIMIT 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to check if auth_sessions table exists: %w", err)
	}

	return &PostgresAuthSessionDB{db: db}, nil
}

func (s *PostgresAuthSessionDB) CreateOrUpdateAuthSession(ctx context.Context, session *model.PersistentAuthSession) error {
	// First check if the session exists
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM auth_sessions 
			WHERE tenant = $1 AND realm = $2 AND session_id_hash = $3
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
				run_id = $1,
				created_at = $2,
				expires_at = $3,
				session_information = $4
			WHERE tenant = $5 AND realm = $6 AND session_id_hash = $7
		`
		_, err = s.db.Exec(ctx, query,
			session.RunID,
			session.CreatedAt,
			session.ExpiresAt,
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
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err = s.db.Exec(ctx, query,
			session.Tenant,
			session.Realm,
			session.RunID,
			session.SessionIDHash,
			session.CreatedAt,
			session.ExpiresAt,
			session.SessionInformation,
		)
		if err != nil {
			return fmt.Errorf("failed to create auth session: %w", err)
		}
	}

	return nil
}

func (s *PostgresAuthSessionDB) GetAuthSessionByID(ctx context.Context, tenant, realm, runID string) (*model.PersistentAuthSession, error) {
	query := `
		SELECT tenant, realm, run_id, session_id_hash,
		       created_at, expires_at, session_information
		FROM auth_sessions
		WHERE tenant = $1 AND realm = $2 AND run_id = $3
	`

	var session model.PersistentAuthSession
	err := s.db.QueryRow(ctx, query, tenant, realm, runID).Scan(
		&session.Tenant,
		&session.Realm,
		&session.RunID,
		&session.SessionIDHash,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.SessionInformation,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get auth session: %w", err)
	}

	return &session, nil
}

func (s *PostgresAuthSessionDB) GetAuthSessionByHash(ctx context.Context, tenant, realm, sessionIDHash string) (*model.PersistentAuthSession, error) {
	query := `
		SELECT tenant, realm, run_id, session_id_hash,
		       created_at, expires_at, session_information
		FROM auth_sessions
		WHERE tenant = $1 AND realm = $2 AND session_id_hash = $3
	`

	var session model.PersistentAuthSession
	err := s.db.QueryRow(ctx, query, tenant, realm, sessionIDHash).Scan(
		&session.Tenant,
		&session.Realm,
		&session.RunID,
		&session.SessionIDHash,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.SessionInformation,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get auth session by hash: %w", err)
	}

	return &session, nil
}

func (s *PostgresAuthSessionDB) ListAuthSessions(ctx context.Context, tenant, realm string) ([]model.PersistentAuthSession, error) {
	query := `
		SELECT tenant, realm, run_id, session_id_hash,
		       created_at, expires_at, session_information
		FROM auth_sessions
		WHERE tenant = $1 AND realm = $2
	`

	rows, err := s.db.Query(ctx, query, tenant, realm)
	if err != nil {
		return nil, fmt.Errorf("failed to list auth sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.PersistentAuthSession
	for rows.Next() {
		var session model.PersistentAuthSession
		err := rows.Scan(
			&session.Tenant,
			&session.Realm,
			&session.RunID,
			&session.SessionIDHash,
			&session.CreatedAt,
			&session.ExpiresAt,
			&session.SessionInformation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan auth session: %w", err)
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s *PostgresAuthSessionDB) ListAllAuthSessions(ctx context.Context, tenant string) ([]model.PersistentAuthSession, error) {
	query := `
		SELECT tenant, realm, run_id, session_id_hash,
		       created_at, expires_at, session_information
		FROM auth_sessions
		WHERE tenant = $1
	`

	rows, err := s.db.Query(ctx, query, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to list all auth sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.PersistentAuthSession
	for rows.Next() {
		var session model.PersistentAuthSession
		err := rows.Scan(
			&session.Tenant,
			&session.Realm,
			&session.RunID,
			&session.SessionIDHash,
			&session.CreatedAt,
			&session.ExpiresAt,
			&session.SessionInformation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan auth session: %w", err)
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s *PostgresAuthSessionDB) DeleteAuthSession(ctx context.Context, tenant, realm, sessionIDHash string) error {
	query := `
		DELETE FROM auth_sessions
		WHERE tenant = $1 AND realm = $2 AND session_id_hash = $3
	`

	_, err := s.db.Exec(ctx, query, tenant, realm, sessionIDHash)
	if err != nil {
		return fmt.Errorf("failed to delete auth session: %w", err)
	}

	return nil
}

func (s *PostgresAuthSessionDB) DeleteExpiredAuthSessions(ctx context.Context, tenant, realm string) error {
	query := `
		DELETE FROM auth_sessions
		WHERE tenant = $1 AND realm = $2 AND expires_at < NOW()
	`

	_, err := s.db.Exec(ctx, query, tenant, realm)
	if err != nil {
		return fmt.Errorf("failed to delete expired auth sessions: %w", err)
	}

	return nil
}
