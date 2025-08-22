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

	// Now get the user from the users table
	// We need to use the existing user DB implementation
	// For now, we'll do a direct query here
	row := p.db.QueryRow(ctx, `
		SELECT id, tenant, realm, username,
		       status,
		       display_name, given_name, family_name,
		       profile_picture_uri,
		       email, phone, email_verified, phone_verified,
		       login_identifier,
		       locale,
		       password_credential, webauthn_credential, mfa_credential,
		       password_locked, webauthn_locked, mfa_locked,
		       failed_login_attempts_password, failed_login_attempts_webauthn, failed_login_attempts_mfa,
		       roles, groups, entitlements, consent, attributes,
		       created_at, updated_at, last_login_at,
		       federated_idp, federated_id,
		       trusted_devices
		FROM users 
		WHERE tenant = $1 AND realm = $2 AND id = $3
	`, tenant, realm, userID)

	// Reuse the existing scan function from user_db.go
	tempUserDB := &PostgresUserDB{db: p.db}
	return tempUserDB.scanUserFromRow(row)
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
