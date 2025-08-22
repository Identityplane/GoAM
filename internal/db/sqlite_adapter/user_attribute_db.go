package sqlite_adapter

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/google/uuid"
)

// SQLiteUserAttributeDB implements the UserAttributeDB interface using SQLite
type SQLiteUserAttributeDB struct {
	db *sql.DB
}

// NewUserAttributeDB creates a new SQLiteUserAttributeDB instance
func NewUserAttributeDB(db *sql.DB) (*SQLiteUserAttributeDB, error) {
	// Check if the connection works and user_attributes table exists by executing a query
	_, err := db.Exec(`
		SELECT 1 FROM user_attributes LIMIT 1
	`)
	if err != nil {
		log := logger.GetLogger()
		log.Debug().Err(err).Msg("warning: failed to check if user_attributes table exists")
	}

	return &SQLiteUserAttributeDB{db: db}, nil
}

func (s *SQLiteUserAttributeDB) CreateUserAttribute(ctx context.Context, attribute model.UserAttribute) error {
	if attribute.ID == "" {
		attribute.ID = uuid.NewString()
	}
	now := time.Now()
	attribute.CreatedAt = now
	attribute.UpdatedAt = now

	// Convert JSON value to string
	valueJSON, err := json.Marshal(attribute.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal attribute value: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO user_attributes (
			id, user_id, tenant, realm, index_value, type, value,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		attribute.ID,
		attribute.UserID,
		attribute.Tenant,
		attribute.Realm,
		attribute.Index,
		attribute.Type,
		string(valueJSON),
		attribute.CreatedAt.Format(time.RFC3339),
		attribute.UpdatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteUserAttributeDB) ListUserAttributes(ctx context.Context, tenant, realm, userID string) ([]model.UserAttribute, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, tenant, realm, index_value, type, value,
		       created_at, updated_at
		FROM user_attributes 
		WHERE tenant = ? AND realm = ? AND user_id = ?
	`, tenant, realm, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize with empty slice instead of nil slice
	attributes := make([]model.UserAttribute, 0)
	for rows.Next() {
		attr, err := s.scanUserAttributeFromRow(rows)
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

func (s *SQLiteUserAttributeDB) GetUserAttributeByID(ctx context.Context, tenant, realm, attributeID string) (*model.UserAttribute, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, tenant, realm, index_value, type, value,
		       created_at, updated_at
		FROM user_attributes 
		WHERE tenant = ? AND realm = ? AND id = ?
	`, tenant, realm, attributeID)

	return s.scanUserAttributeFromRow(row)
}

func (s *SQLiteUserAttributeDB) UpdateUserAttribute(ctx context.Context, attribute *model.UserAttribute) error {
	attribute.UpdatedAt = time.Now()

	// Convert JSON value to string
	valueJSON, err := json.Marshal(attribute.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal attribute value: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE user_attributes SET
			index_value = ?,
			type = ?,
			value = ?,
			updated_at = ?
		WHERE tenant = ? AND realm = ? AND id = ?
	`,
		attribute.Index,
		attribute.Type,
		string(valueJSON),
		attribute.UpdatedAt.Format(time.RFC3339),
		attribute.Tenant,
		attribute.Realm,
		attribute.ID,
	)

	return err
}

func (s *SQLiteUserAttributeDB) DeleteUserAttribute(ctx context.Context, tenant, realm, attributeID string) error {
	query := `
		DELETE FROM user_attributes 
		WHERE tenant = ? AND realm = ? AND id = ?
	`

	_, err := s.db.ExecContext(ctx, query, tenant, realm, attributeID)
	if err != nil {
		return fmt.Errorf("failed to delete user attribute: %w", err)
	}

	// No error if attribute doesn't exist (idempotent)
	return nil
}

func (s *SQLiteUserAttributeDB) GetUserByAttributeIndex(ctx context.Context, tenant, realm, attributeType, index string) (*model.User, error) {
	// First get the user_id from the attribute
	var userID string
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id FROM user_attributes 
		WHERE tenant = ? AND realm = ? AND type = ? AND index_value = ?
	`, tenant, realm, attributeType, index).Scan(&userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	// Now get the user from the users table
	// We need to use the existing user DB implementation
	// For now, we'll do a direct query here
	row := s.db.QueryRowContext(ctx, `
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
		WHERE tenant = ? AND realm = ? AND id = ?
	`, tenant, realm, userID)

	// Reuse the existing scan function from user_db.go
	// We'll need to create a temporary SQLiteUserDB instance to use its scanUserFromRow method
	tempUserDB := &SQLiteUserDB{db: s.db}
	return tempUserDB.scanUserFromRow(row)
}

// scanUserAttributeFromRow scans a user attribute from a database row
func (s *SQLiteUserAttributeDB) scanUserAttributeFromRow(scanner interface{}) (*model.UserAttribute, error) {
	var attribute model.UserAttribute
	var valueJSON string
	var createdAt, updatedAt string

	var err error
	switch r := scanner.(type) {
	case *sql.Row:
		err = r.Scan(
			&attribute.ID,
			&attribute.UserID,
			&attribute.Tenant,
			&attribute.Realm,
			&attribute.Index,
			&attribute.Type,
			&valueJSON,
			&createdAt,
			&updatedAt,
		)
	case *sql.Rows:
		err = r.Scan(
			&attribute.ID,
			&attribute.UserID,
			&attribute.Tenant,
			&attribute.Realm,
			&attribute.Index,
			&attribute.Type,
			&valueJSON,
			&createdAt,
			&updatedAt,
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
	attribute.CreatedAt = createdAtTime.Local()
	attribute.UpdatedAt = updatedAtTime.Local()

	// Parse JSON value
	if valueJSON != "" && valueJSON != "null" {
		// We'll store the raw JSON as interface{} for now
		// The application layer can handle the specific type conversion
		var rawValue interface{}
		if err := json.Unmarshal([]byte(valueJSON), &rawValue); err != nil {
			return nil, fmt.Errorf("failed to unmarshal attribute value: %w", err)
		}
		attribute.Value = rawValue
	}

	return &attribute, nil
}
