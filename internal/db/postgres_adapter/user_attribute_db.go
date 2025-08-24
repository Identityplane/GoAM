package postgres_adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresUserAttributeDB implements the UserAttributeDB interface using PostgreSQL
type PostgresUserAttributeDB struct {
	db *pgxpool.Pool
}

// NewPostgresUserAttributeDB creates a new PostgresUserAttributeDB instance
func NewPostgresUserAttributeDB(db *pgxpool.Pool) (*PostgresUserAttributeDB, error) {
	// Check if the connection works and user_attributes table exists by executing a query
	_, err := db.Exec(context.Background(), `
		SELECT 1 FROM user_attributes LIMIT 1
	`)
	if err != nil {
		log := logger.GetLogger()
		log.Debug().Err(err).Msg("warning: failed to check if user_attributes table exists")
	}

	return &PostgresUserAttributeDB{db: db}, nil
}

func (p *PostgresUserAttributeDB) CreateUserAttribute(ctx context.Context, attribute model.UserAttribute) error {
	if attribute.ID == "" {
		attribute.ID = uuid.NewString()
	}
	now := time.Now()
	attribute.CreatedAt = now
	attribute.UpdatedAt = now

	// Convert value to JSONB - PostgreSQL will handle the JSON conversion natively
	valueJSONB, err := json.Marshal(attribute.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal attribute value: %w", err)
	}

	_, err = p.db.Exec(ctx, `
		INSERT INTO user_attributes (
			id, user_id, tenant, realm, index_value, type, value,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		attribute.ID,
		attribute.UserID,
		attribute.Tenant,
		attribute.Realm,
		attribute.Index,
		attribute.Type,
		valueJSONB, // This will be stored as JSONB
		attribute.CreatedAt,
		attribute.UpdatedAt,
	)

	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresUserAttributeDB) ListUserAttributes(ctx context.Context, tenant, realm, userID string) ([]model.UserAttribute, error) {
	rows, err := p.db.Query(ctx, `
		SELECT id, user_id, tenant, realm, index_value, type, value,
		       created_at, updated_at
		FROM user_attributes 
		WHERE tenant = $1 AND realm = $2 AND user_id = $3
	`, tenant, realm, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize with empty slice instead of nil slice
	attributes := make([]model.UserAttribute, 0)
	for rows.Next() {
		attr, err := p.scanUserAttributeFromRow(rows)
		if err != nil {
			return nil, err
		}
		attributes = append(attributes, *attr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return attributes, nil
}

func (p *PostgresUserAttributeDB) GetUserAttributeByID(ctx context.Context, tenant, realm, attributeID string) (*model.UserAttribute, error) {
	row := p.db.QueryRow(ctx, `
		SELECT id, user_id, tenant, realm, index_value, type, value,
		       created_at, updated_at
		FROM user_attributes 
		WHERE tenant = $1 AND realm = $2 AND id = $3
	`, tenant, realm, attributeID)

	return p.scanUserAttributeFromRow(row)
}

func (p *PostgresUserAttributeDB) UpdateUserAttribute(ctx context.Context, attribute *model.UserAttribute) error {
	attribute.UpdatedAt = time.Now()

	// Convert value to JSONB
	valueJSONB, err := json.Marshal(attribute.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal attribute value: %w", err)
	}

	_, err = p.db.Exec(ctx, `
		UPDATE user_attributes SET
			index_value = $1,
			type = $2,
			value = $3,
			updated_at = $4
		WHERE tenant = $5 AND realm = $6 AND id = $7
	`,
		attribute.Index,
		attribute.Type,
		valueJSONB, // This will be stored as JSONB
		attribute.UpdatedAt,
		attribute.Tenant,
		attribute.Realm,
		attribute.ID,
	)

	return err
}

func (p *PostgresUserAttributeDB) DeleteUserAttribute(ctx context.Context, tenant, realm, attributeID string) error {
	query := `
		DELETE FROM user_attributes 
		WHERE tenant = $1 AND realm = $2 AND id = $3
	`

	_, err := p.db.Exec(ctx, query, tenant, realm, attributeID)
	if err != nil {
		return fmt.Errorf("failed to delete user attribute: %w", err)
	}

	// No error if attribute doesn't exist (idempotent)
	return nil
}

func (p *PostgresUserAttributeDB) GetUserByAttributeIndex(ctx context.Context, tenant, realm, attributeType, index string) (*model.User, error) {
	// First get the user_id from the attribute
	var userID string
	err := p.db.QueryRow(ctx, `
		SELECT user_id FROM user_attributes 
		WHERE tenant = $1 AND realm = $2 AND type = $3 AND index_value = $4
	`, tenant, realm, attributeType, index).Scan(&userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	// Now get the user from the users table using the correct columns
	row := p.db.QueryRow(ctx, `
		SELECT id, tenant, realm, status,
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE tenant = $1 AND realm = $2 AND id = $3
	`, tenant, realm, userID)

	// Reuse the existing scan function from user_db.go
	tempUserDB := &PostgresUserDB{db: p.db}
	return tempUserDB.scanUserFromRow(row)
}

// GetUserWithAttributes loads a user with all their attributes in one efficient database query
func (p *PostgresUserAttributeDB) GetUserWithAttributes(ctx context.Context, tenant, realm, userID string) (*model.User, error) {
	// Single query to get user with attributes as JSONB using simplified schema
	row := p.db.QueryRow(ctx, `
		SELECT 
			u.id, u.tenant, u.realm, u.status,
			u.created_at, u.updated_at, u.last_login_at,
			COALESCE(
				(
					SELECT jsonb_agg(
						jsonb_build_object(
							'id', ua.id,
							'user_id', ua.user_id,
							'tenant', ua.tenant,
							'realm', ua.realm,
							'index', ua.index_value,
							'type', ua.type,
							'value', ua.value,
							'created_at', ua.created_at,
							'updated_at', ua.updated_at
						)
					)
					FROM user_attributes ua
					WHERE ua.tenant = $1 AND ua.realm = $2 AND ua.user_id = $3
				),
				'[]'::jsonb
			) as user_attributes
		FROM users u
		WHERE u.tenant = $1 AND u.realm = $2 AND u.id = $3
	`, tenant, realm, userID)

	// Scan the result including the JSONB attributes
	var user model.User
	var userAttributesJSONB []byte
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
		&userAttributesJSONB,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("failed to scan user with attributes: %w", err)
	}

	// Set timestamps
	user.CreatedAt = createdAt.UTC()
	user.UpdatedAt = updatedAt.UTC()
	user.LastLoginAt = lastLoginAt

	// Parse the JSONB attributes
	if len(userAttributesJSONB) > 0 {
		var attributes []model.UserAttribute
		if err := json.Unmarshal(userAttributesJSONB, &attributes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user attributes: %w", err)
		}
		user.UserAttributes = attributes
	} else {
		user.UserAttributes = []model.UserAttribute{}
	}

	return &user, nil
}

// GetUserByAttributeIndexWithAttributes finds a user by attribute index and loads all their attributes
// Uses a single efficient query with JOIN and subquery for production performance
func (p *PostgresUserAttributeDB) GetUserByAttributeIndexWithAttributes(ctx context.Context, tenant, realm, attributeType, index string) (*model.User, error) {
	// Single query with JOIN to get user by attribute index and all their attributes using simplified schema
	row := p.db.QueryRow(ctx, `
		SELECT 
			u.id, u.tenant, u.realm, u.status,
			u.created_at, u.updated_at, u.last_login_at,
			COALESCE(
				(
					SELECT jsonb_agg(
						jsonb_build_object(
							'id', ua.id,
							'user_id', ua.user_id,
							'tenant', ua.tenant,
							'realm', ua.realm,
							'index', ua.index_value,
							'type', ua.type,
							'value', ua.value,
							'created_at', ua.created_at,
							'updated_at', ua.updated_at
						)
					)
					FROM user_attributes ua
					WHERE ua.tenant = $1 AND ua.realm = $2 AND ua.user_id = u.id
				),
				'[]'::jsonb
			) as user_attributes
		FROM users u
		INNER JOIN user_attributes ua_lookup ON u.id = ua_lookup.user_id
		WHERE u.tenant = $1 
		  AND u.realm = $2 
		  AND ua_lookup.type = $3 
		  AND ua_lookup.index_value = $4
	`, tenant, realm, attributeType, index)

	// Scan the result including the JSONB attributes
	var user model.User
	var userAttributesJSONB []byte
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
		&userAttributesJSONB,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("failed to scan user with attributes by index: %w", err)
	}

	// Set timestamps
	user.CreatedAt = createdAt.UTC()
	user.UpdatedAt = updatedAt.UTC()
	user.LastLoginAt = lastLoginAt

	// Parse the JSONB attributes
	if len(userAttributesJSONB) > 0 {
		var attributes []model.UserAttribute
		if err := json.Unmarshal(userAttributesJSONB, &attributes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user attributes: %w", err)
		}
		user.UserAttributes = attributes
	} else {
		user.UserAttributes = []model.UserAttribute{}
	}

	return &user, nil
}

// CreateUserWithAttributes creates a user and all their attributes in a single transaction
func (p *PostgresUserAttributeDB) CreateUserWithAttributes(ctx context.Context, user *model.User) error {
	// Start a transaction
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Generate user ID if not set
	if user.ID == "" {
		user.ID = uuid.NewString()
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Create the user first
	var lastLoginAt interface{}
	if user.LastLoginAt != nil {
		lastLoginAt = user.LastLoginAt
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO users (
			id, tenant, realm, status,
			created_at, updated_at, last_login_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		user.ID,
		user.Tenant,
		user.Realm,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
		lastLoginAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Create each attribute
	for i := range user.UserAttributes {
		attribute := &user.UserAttributes[i]

		// Set the user_id if not already set
		if attribute.UserID == "" {
			attribute.UserID = user.ID
		}

		// Set tenant and realm if not already set
		if attribute.Tenant == "" {
			attribute.Tenant = user.Tenant
		}
		if attribute.Realm == "" {
			attribute.Realm = user.Realm
		}

		// Generate ID if not set
		if attribute.ID == "" {
			attribute.ID = uuid.NewString()
		}

		// Set timestamps
		now := time.Now()
		attribute.CreatedAt = now
		attribute.UpdatedAt = now

		// Convert value to JSONB - PostgreSQL will handle the JSON conversion natively
		valueJSONB, err := json.Marshal(attribute.Value)
		if err != nil {
			return fmt.Errorf("failed to marshal attribute value: %w", err)
		}

		// Insert the attribute
		_, err = tx.Exec(ctx, `
			INSERT INTO user_attributes (
				id, user_id, tenant, realm, index_value, type, value,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
			attribute.ID,
			attribute.UserID,
			attribute.Tenant,
			attribute.Realm,
			attribute.Index,
			attribute.Type,
			valueJSONB,
			attribute.CreatedAt,
			attribute.UpdatedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to create attribute %s: %w", attribute.Type, err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// scanUserAttributeFromRow scans a user attribute from a database row
func (p *PostgresUserAttributeDB) scanUserAttributeFromRow(scanner interface{}) (*model.UserAttribute, error) {
	var attribute model.UserAttribute
	var valueJSONB []byte
	var createdAt, updatedAt time.Time

	var err error
	switch r := scanner.(type) {
	case pgx.Row:
		err = r.Scan(
			&attribute.ID,
			&attribute.UserID,
			&attribute.Tenant,
			&attribute.Realm,
			&attribute.Index,
			&attribute.Type,
			&valueJSONB,
			&createdAt,
			&updatedAt,
		)
	case pgx.Rows:
		err = r.Scan(
			&attribute.ID,
			&attribute.UserID,
			&attribute.Tenant,
			&attribute.Realm,
			&attribute.Index,
			&attribute.Type,
			&valueJSONB,
			&createdAt,
			&updatedAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type: %T", scanner)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	// Set timestamps with timezone information preserved
	attribute.CreatedAt = createdAt.UTC()
	attribute.UpdatedAt = updatedAt.UTC()

	// Parse JSONB value - PostgreSQL returns it as []byte
	if len(valueJSONB) > 0 {
		// We'll store the raw JSON as interface{} for now
		// The application layer can handle the specific type conversion
		var rawValue interface{}
		if err := json.Unmarshal(valueJSONB, &rawValue); err != nil {
			return nil, fmt.Errorf("failed to unmarshal attribute value: %w", err)
		}
		attribute.Value = rawValue
	}

	return &attribute, nil
}
