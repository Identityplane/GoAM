package sqlite_adapter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/google/uuid"
)

// SQLiteUserDB implements the UserDB interface using SQLite
type SQLiteUserDB struct {
	db *sql.DB
}

// NewUserDB creates a new SQLiteUserDB instance
func NewUserDB(db *sql.DB) (*SQLiteUserDB, error) {
	// Check if the connection works and users table exists by executing a query
	_, err := db.Exec(`
		SELECT 1 FROM users LIMIT 1
	`)
	if err != nil {
		log := logger.GetGoamLogger()
		log.Debug().Err(err).Msg("warning: failed to check if users table exists")
	}

	return &SQLiteUserDB{db: db}, nil
}

func (s *SQLiteUserDB) CreateUser(ctx context.Context, user model.User) error {
	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Handle time fields
	var lastLoginAt interface{}
	if user.LastLoginAt != nil {
		lastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (
			id, tenant, realm, status,
			created_at, updated_at, last_login_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		user.ID,
		user.Tenant,
		user.Realm,
		user.Status,
		user.CreatedAt.Format(time.RFC3339),
		user.UpdatedAt.Format(time.RFC3339),
		lastLoginAt,
	)

	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteUserDB) GetUserByID(ctx context.Context, tenant, realm, userID string) (*model.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, tenant, realm, status,
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE tenant = ? AND realm = ? AND id = ?
	`, tenant, realm, userID)

	return s.scanUserFromRow(row)
}

// scanUserFromRow scans a user from a database row
func (s *SQLiteUserDB) scanUserFromRow(scanner interface{}) (*model.User, error) {
	var user model.User
	var createdAt, updatedAt string
	var lastLoginAt sql.NullString

	var err error
	switch r := scanner.(type) {
	case *sql.Row:
		err = r.Scan(
			&user.ID,
			&user.Tenant,
			&user.Realm,
			&user.Status,
			&createdAt,
			&updatedAt,
			&lastLoginAt,
		)
	case *sql.Rows:
		err = r.Scan(
			&user.ID,
			&user.Tenant,
			&user.Realm,
			&user.Status,
			&createdAt,
			&updatedAt,
			&lastLoginAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type: %T", scanner)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	// Parse timestamps
	createdAtTime, _ := time.Parse(time.RFC3339, createdAt)
	updatedAtTime, _ := time.Parse(time.RFC3339, updatedAt)

	// Convert to local time to match PostgreSQL behavior
	user.CreatedAt = createdAtTime.Local()
	user.UpdatedAt = updatedAtTime.Local()

	if lastLoginAt.Valid {
		lastLogin, _ := time.Parse(time.RFC3339, lastLoginAt.String)
		lastLoginLocal := lastLogin.Local()
		user.LastLoginAt = &lastLoginLocal
	}

	return &user, nil
}

func (s *SQLiteUserDB) UpdateUser(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()

	// Handle time fields
	var lastLoginAt interface{}
	if user.LastLoginAt != nil {
		lastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}

	result, err := s.db.ExecContext(ctx, `
		UPDATE users SET
			status = ?,
			updated_at = ?,
			last_login_at = ?
		WHERE tenant = ? AND realm = ? AND id = ?
	`,
		user.Status,
		user.UpdatedAt.Format(time.RFC3339),
		lastLoginAt,
		user.Tenant,
		user.Realm,
		user.ID,
	)

	if err != nil {
		return err
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: tenant=%s, realm=%s, id=%s", user.Tenant, user.Realm, user.ID)
	}

	return nil
}

func (s *SQLiteUserDB) GetUserStats(ctx context.Context, tenant, realm string) (*model.UserStats, error) {
	var stats model.UserStats

	// TODO implement this with attributes

	// Query to get all user statistics in a single query
	err := s.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_users,
			COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive_users,
			COUNT(CASE WHEN status = 'locked' THEN 1 END) as locked_users
		FROM users 
		WHERE tenant = ? AND realm = ?
	`, tenant, realm).Scan(
		&stats.TotalUsers,
		&stats.ActiveUsers,
		&stats.InactiveUsers,
		&stats.LockedUsers,
	)

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (s *SQLiteUserDB) CountUsers(ctx context.Context, tenant, realm string) (int64, error) {
	var count int64
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM users WHERE tenant = ? AND realm = ?
	`, tenant, realm).Scan(&count)
	return count, err
}

// ListUsers implements model.UserDB.
func (s *SQLiteUserDB) ListUsers(ctx context.Context, tenant string, realm string) ([]model.User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant, realm, status,
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE tenant = ? AND realm = ?
	`, tenant, realm)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		user, err := s.scanUserFromRow(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *SQLiteUserDB) ListUsersWithPagination(ctx context.Context, tenant, realm string, offset, limit int) ([]model.User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant, realm, status,
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE tenant = ? AND realm = ?
		LIMIT ? OFFSET ?
	`, tenant, realm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		user, err := s.scanUserFromRow(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// DeleteUser deletes a user by userID
func (s *SQLiteUserDB) DeleteUser(ctx context.Context, tenant, realm, userID string) error {
	query := `
		DELETE FROM users 
		WHERE tenant = ? AND realm = ? AND id = ?
	`

	_, err := s.db.ExecContext(ctx, query, tenant, realm, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// No error if user doesn't exist (idempotent)
	return nil
}

func (s *SQLiteUserDB) GetUserByEmail(ctx context.Context, tenant, realm, email string) (*model.User, error) {
	// Note: This method needs to be implemented by joining with user_attributes table
	// since email is stored as a user attribute, not directly on the users table
	// For now, return an error indicating this needs to be implemented
	return nil, fmt.Errorf("GetUserByEmail not implemented - email is stored as user attribute")
}

func (s *SQLiteUserDB) GetUserByLoginIdentifier(ctx context.Context, tenant, realm, loginIdentifier string) (*model.User, error) {
	// Note: This method needs to be implemented by joining with user_attributes table
	// since login_identifier is stored as a user attribute, not directly on the users table
	// For now, return an error indicating this needs to be implemented
	return nil, fmt.Errorf("GetUserByLoginIdentifier not implemented - login_identifier is stored as user attribute")
}

func (s *SQLiteUserDB) GetUserByFederatedIdentifier(ctx context.Context, tenant, realm, provider, identifier string) (*model.User, error) {
	// Note: This method needs to be implemented by joining with user_attributes table
	// since federated identifiers are stored as user attributes, not directly on the users table
	// For now, return an error indicating this needs to be implemented
	return nil, fmt.Errorf("GetUserByFederatedIdentifier not implemented - federated identifiers are stored as user attributes")
}
