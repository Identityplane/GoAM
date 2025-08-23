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

// GetUserWithAttributes loads a user with all their attributes in one database query
func (s *SQLiteUserAttributeDB) GetUserWithAttributes(ctx context.Context, tenant, realm, userID string) (*model.User, error) {
	// First get the user
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

	tempUserDB := &SQLiteUserDB{db: s.db}
	user, err := tempUserDB.scanUserFromRow(row)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	// Then get all attributes for this user
	attributes, err := s.ListUserAttributes(ctx, tenant, realm, userID)
	if err != nil {
		return nil, err
	}

	user.UserAttributes = attributes
	return user, nil
}

// GetUserByAttributeIndexWithAttributes finds a user by attribute index and loads all their attributes
func (s *SQLiteUserAttributeDB) GetUserByAttributeIndexWithAttributes(ctx context.Context, tenant, realm, attributeType, index string) (*model.User, error) {
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

	// Now get the user with all attributes
	return s.GetUserWithAttributes(ctx, tenant, realm, userID)
}

// CreateUserWithAttributes creates a user and all their attributes in a single transaction
func (s *SQLiteUserAttributeDB) CreateUserWithAttributes(ctx context.Context, user *model.User) error {
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
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

	// Convert JSON fields to strings for SQLite
	rolesJSON, _ := json.Marshal(user.Roles)
	groupsJSON, _ := json.Marshal(user.Groups)
	attributesJSON, _ := json.Marshal(user.Attributes)
	trustedDevicesJSON, _ := json.Marshal(user.TrustedDevices)
	entitlementsJSON, _ := json.Marshal(user.Entitlements)
	consentJSON, _ := json.Marshal(user.Consent)

	// Handle time fields
	var updatedAt, lastLoginAt interface{}
	if !user.UpdatedAt.IsZero() {
		updatedAt = user.UpdatedAt.Format(time.RFC3339)
	}
	if user.LastLoginAt != nil {
		lastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}

	// Create the user first
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (
			id, tenant, realm, username,
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
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		user.ID,
		user.Tenant,
		user.Realm,
		user.Username,
		user.Status,
		user.DisplayName,
		user.GivenName,
		user.FamilyName,
		user.ProfilePictureURI,
		user.Email,
		user.Phone,
		user.EmailVerified,
		user.PhoneVerified,
		user.LoginIdentifier,
		user.Locale,
		user.PasswordCredential,
		user.WebAuthnCredential,
		user.MFACredential,
		user.PasswordLocked,
		user.WebAuthnLocked,
		user.MFALocked,
		user.FailedLoginAttemptsPassword,
		user.FailedLoginAttemptsWebAuthn,
		user.FailedLoginAttemptsMFA,
		string(rolesJSON),
		string(groupsJSON),
		string(entitlementsJSON),
		string(consentJSON),
		string(attributesJSON),
		user.CreatedAt.Format(time.RFC3339),
		updatedAt,
		lastLoginAt,
		user.FederatedIDP,
		user.FederatedID,
		string(trustedDevicesJSON),
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

		// Convert JSON value to string
		valueJSON, err := json.Marshal(attribute.Value)
		if err != nil {
			return fmt.Errorf("failed to marshal attribute value: %w", err)
		}

		// Insert the attribute
		_, err = tx.ExecContext(ctx, `
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
			return fmt.Errorf("failed to create attribute %s: %w", attribute.Type, err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
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
