package postgres_adapter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresUserDB implements the UserDB interface using PostgreSQL
type PostgresUserDB struct {
	db *pgxpool.Pool
}

// NewPostgresUserDB creates a new PostgresUserDB instance
func NewPostgresUserDB(db *pgxpool.Pool) (*PostgresUserDB, error) {
	// Check if the connection works and users table exists by executing a query
	_, err := db.Exec(context.Background(), `
		SELECT 1 FROM users LIMIT 1
	`)
	if err != nil {
		log := logger.GetGoamLogger()
		log.Debug().Err(err).Msg("warning: failed to check if users table exists")
	}

	return &PostgresUserDB{db: db}, nil
}

func (p *PostgresUserDB) CreateUser(ctx context.Context, user model.User) error {
	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Handle time fields
	var lastLoginAt interface{}
	if user.LastLoginAt != nil {
		lastLoginAt = user.LastLoginAt
	}

	_, err := p.db.Exec(ctx, `
		INSERT INTO users (
			id, tenant, realm, status,
			created_at, updated_at, last_login_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		user.ID,        // $1
		user.Tenant,    // $2
		user.Realm,     // $3
		user.Status,    // $4
		user.CreatedAt, // $5
		user.UpdatedAt, // $6
		lastLoginAt,    // $7
	)

	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresUserDB) UpdateUser(ctx context.Context, user *model.User) error {
	if user.ID == "" {
		return fmt.Errorf("user ID is required")
	}

	user.UpdatedAt = time.Now()

	// Handle time fields
	var lastLoginAt interface{}
	if user.LastLoginAt != nil {
		lastLoginAt = user.LastLoginAt
	}

	result, err := p.db.Exec(ctx, `
		UPDATE users SET
			status = $1,
			updated_at = $2,
			last_login_at = $3
		WHERE id = $4 AND tenant = $5 AND realm = $6
	`,
		user.Status,
		user.UpdatedAt,
		lastLoginAt,
		user.ID,
		user.Tenant,
		user.Realm,
	)

	if err != nil {
		return err
	}

	// Check if any rows were affected
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found: tenant=%s, realm=%s, id=%s", user.Tenant, user.Realm, user.ID)
	}

	return nil
}

func (p *PostgresUserDB) ListUsers(ctx context.Context, tenant, realm string) ([]model.User, error) {
	rows, err := p.db.Query(ctx, `
		SELECT id, tenant, realm, status, created_at, updated_at, last_login_at
		FROM users 
		WHERE tenant = $1 AND realm = $2
	`, tenant, realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		var createdAt, updatedAt time.Time
		var lastLoginAt *time.Time

		err := rows.Scan(
			&user.ID,
			&user.Tenant,
			&user.Realm,
			&user.Status,
			&createdAt,
			&updatedAt,
			&lastLoginAt,
		)
		if err != nil {
			return nil, err
		}

		// Set timestamps with timezone information preserved
		user.CreatedAt = createdAt.UTC()
		user.UpdatedAt = updatedAt.UTC()
		if lastLoginAt != nil {
			lastLogin := lastLoginAt.UTC()
			user.LastLoginAt = &lastLogin
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (p *PostgresUserDB) ListUsersWithPagination(ctx context.Context, tenant, realm string, offset, limit int) ([]model.User, error) {
	rows, err := p.db.Query(ctx, `
		SELECT id, tenant, realm, status, created_at, updated_at, last_login_at
		FROM users 
		WHERE tenant = $1 AND realm = $2
		LIMIT $3 OFFSET $4
	`, tenant, realm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		var createdAt, updatedAt time.Time
		var lastLoginAt *time.Time

		err := rows.Scan(
			&user.ID,
			&user.Tenant,
			&user.Realm,
			&user.Status,
			&createdAt,
			&updatedAt,
			&lastLoginAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		// Set timestamps with timezone information preserved
		user.CreatedAt = createdAt.UTC()
		user.UpdatedAt = updatedAt.UTC()
		if lastLoginAt != nil {
			lastLogin := lastLoginAt.UTC()
			user.LastLoginAt = &lastLogin
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (p *PostgresUserDB) CountUsers(ctx context.Context, tenant, realm string) (int64, error) {
	var count int64
	err := p.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM users WHERE tenant = $1 AND realm = $2
	`, tenant, realm).Scan(&count)
	return count, err
}

func (p *PostgresUserDB) GetUserStats(ctx context.Context, tenant, realm string) (*model.UserStats, error) {
	var stats model.UserStats

	// TODO implement this with attributes

	// Query to get all user statistics in a single query
	err := p.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_users,
			COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive_users,
			COUNT(CASE WHEN status = 'locked' THEN 1 END) as locked_users
		FROM users 
		WHERE tenant = $1 AND realm = $2
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

// DeleteUser deletes a user by userID
func (p *PostgresUserDB) DeleteUser(ctx context.Context, tenant, realm, userID string) error {
	query := `
		DELETE FROM users 
		WHERE tenant = $1 AND realm = $2 AND id = $3
	`

	_, err := p.db.Exec(ctx, query, tenant, realm, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// No error if user doesn't exist (idempotent)
	return nil
}

// scanUserFromRow scans a user from a database row
func (p *PostgresUserDB) scanUserFromRow(row pgx.Row) (*model.User, error) {
	var user model.User
	var createdAt, updatedAt time.Time
	var lastLoginAt *time.Time

	err := row.Scan(
		&user.ID,
		&user.Tenant,
		&user.Realm,
		&user.Status,
		&createdAt,
		&updatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	// Set timestamps with timezone information preserved
	user.CreatedAt = createdAt.UTC()
	user.UpdatedAt = updatedAt.UTC()
	if lastLoginAt != nil {
		lastLogin := lastLoginAt.UTC()
		user.LastLoginAt = &lastLogin
	}

	return &user, nil
}

func (p *PostgresUserDB) GetUserByID(ctx context.Context, tenant, realm, userID string) (*model.User, error) {
	row := p.db.QueryRow(ctx, `
		SELECT id, tenant, realm, status, created_at, updated_at, last_login_at
		FROM users 
		WHERE tenant = $1 AND realm = $2 AND id = $3
	`, tenant, realm, userID)

	return p.scanUserFromRow(row)
}

func (p *PostgresUserDB) GetUserByEmail(ctx context.Context, tenant, realm, email string) (*model.User, error) {
	// Note: This method needs to be implemented by joining with user_attributes table
	// since email is stored as a user attribute, not directly on the users table
	// For now, return an error indicating this needs to be implemented
	return nil, fmt.Errorf("GetUserByEmail not implemented - email is stored as user attribute")
}

func (p *PostgresUserDB) GetUserByLoginIdentifier(ctx context.Context, tenant, realm, loginIdentifier string) (*model.User, error) {
	// Note: This method needs to be implemented by joining with user_attributes table
	// since login_identifier is stored as a user attribute, not directly on the users table
	// For now, return an error indicating this needs to be implemented
	return nil, fmt.Errorf("GetUserByLoginIdentifier not implemented - login_identifier is stored as user attribute")
}

func (p *PostgresUserDB) GetUserByFederatedIdentifier(ctx context.Context, tenant, realm, provider, identifier string) (*model.User, error) {
	// Note: This method needs to be implemented by joining with user_attributes table
	// since federated identifiers are stored as user attributes, not directly on the users table
	// For now, return an error indicating this needs to be implemented
	return nil, fmt.Errorf("GetUserByFederatedIdentifier not implemented - federated identifiers are stored as user attributes")
}
