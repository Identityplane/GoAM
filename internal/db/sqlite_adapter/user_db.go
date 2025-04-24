package sqlite_adapter

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"goiam/internal/logger"
	"goiam/internal/model"
	"time"

	"github.com/google/uuid"
)

// SQLiteUserDB implements the UserDB interface using SQLite
type SQLiteUserDB struct {
	db *sql.DB
}

// ListUsers implements model.UserDB.
func (s *SQLiteUserDB) ListUsers(ctx context.Context, tenant string, realm string) ([]model.User, error) {
	panic("unimplemented")
}

// NewUserDB creates a new SQLiteUserDB instance
func NewUserDB(db *sql.DB) (*SQLiteUserDB, error) {

	// Check if the connection works and users table exists by executing a query
	_, err := db.Exec(`
		SELECT 1 FROM users LIMIT 1
	`)
	if err != nil {
		logger.DebugNoContext("Warning: failed to check if users table exists: %v", err)
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

	// Convert JSON fields to strings
	rolesJSON, _ := json.Marshal(user.Roles)
	groupsJSON, _ := json.Marshal(user.Groups)
	attributesJSON, _ := json.Marshal(user.Attributes)
	trustedDevicesJSON, _ := json.Marshal(user.TrustedDevices)

	// Handle time fields
	var updatedAt, lastLoginAt interface{}
	if !user.UpdatedAt.IsZero() {
		updatedAt = user.UpdatedAt.Format(time.RFC3339)
	}
	if user.LastLoginAt != nil {
		lastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (
			id, tenant, realm, username,
			status,
			display_name, given_name, family_name,
			email, phone, email_verified, phone_verified,
			locale,
			password_credential, webauthn_credential, mfa_credential,
			password_locked, webauthn_locked, mfa_locked,
			failed_login_attempts_password, failed_login_attempts_webauthn, failed_login_attempts_mfa,
			roles, groups, attributes,
			created_at, updated_at, last_login_at,
			federated_idp, federated_id,
			trusted_devices
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		user.ID,
		user.Tenant,
		user.Realm,
		user.Username,
		user.Status,
		user.DisplayName,
		user.GivenName,
		user.FamilyName,
		user.Email,
		user.Phone,
		user.EmailVerified,
		user.PhoneVerified,
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
		string(attributesJSON),
		user.CreatedAt.Format(time.RFC3339),
		updatedAt,
		lastLoginAt,
		user.FederatedIDP,
		user.FederatedID,
		string(trustedDevicesJSON),
	)

	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteUserDB) GetUserByUsername(ctx context.Context, tenant, realm, username string) (*model.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, tenant, realm, username,
		       status,
		       display_name, given_name, family_name,
		       email, phone, email_verified, phone_verified,
		       locale,
		       password_credential, webauthn_credential, mfa_credential,
		       password_locked, webauthn_locked, mfa_locked,
		       failed_login_attempts_password, failed_login_attempts_webauthn, failed_login_attempts_mfa,
		       roles, groups, attributes,
		       created_at, updated_at, last_login_at,
		       federated_idp, federated_id,
		       trusted_devices
		FROM users 
		WHERE tenant = ? AND realm = ? AND username = ?
	`, tenant, realm, username)

	var user model.User
	var rolesJSON, groupsJSON, attributesJSON, trustedDevicesJSON string
	var createdAt, updatedAt string
	var lastLoginAt sql.NullString

	err := row.Scan(
		&user.ID,
		&user.Tenant,
		&user.Realm,
		&user.Username,
		&user.Status,
		&user.DisplayName,
		&user.GivenName,
		&user.FamilyName,
		&user.Email,
		&user.Phone,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.Locale,
		&user.PasswordCredential,
		&user.WebAuthnCredential,
		&user.MFACredential,
		&user.PasswordLocked,
		&user.WebAuthnLocked,
		&user.MFALocked,
		&user.FailedLoginAttemptsPassword,
		&user.FailedLoginAttemptsWebAuthn,
		&user.FailedLoginAttemptsMFA,
		&rolesJSON,
		&groupsJSON,
		&attributesJSON,
		&createdAt,
		&updatedAt,
		&lastLoginAt,
		&user.FederatedIDP,
		&user.FederatedID,
		&trustedDevicesJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	// Parse timestamps
	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	if lastLoginAt.Valid {
		lastLogin, _ := time.Parse(time.RFC3339, lastLoginAt.String)
		user.LastLoginAt = &lastLogin
	}

	// Parse JSON fields
	user.Roles = []string{}
	user.Groups = []string{}
	user.Attributes = map[string]string{}

	if rolesJSON != "" && rolesJSON != "null" {
		_ = json.Unmarshal([]byte(rolesJSON), &user.Roles)
	} else {
		user.Roles = []string{}
	}
	if groupsJSON != "" && groupsJSON != "null" {
		_ = json.Unmarshal([]byte(groupsJSON), &user.Groups)
	} else {
		user.Groups = []string{}
	}
	if attributesJSON != "" && attributesJSON != "null" {
		_ = json.Unmarshal([]byte(attributesJSON), &user.Attributes)
	} else {
		user.Attributes = map[string]string{}
	}

	user.TrustedDevices = trustedDevicesJSON

	return &user, nil
}

func (s *SQLiteUserDB) UpdateUser(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()

	// Convert JSON fields to strings
	rolesJSON, _ := json.Marshal(user.Roles)
	groupsJSON, _ := json.Marshal(user.Groups)
	attributesJSON, _ := json.Marshal(user.Attributes)
	trustedDevicesJSON, _ := json.Marshal(user.TrustedDevices)

	// Handle time fields
	var updatedAt, lastLoginAt interface{}
	if !user.UpdatedAt.IsZero() {
		updatedAt = user.UpdatedAt.Format(time.RFC3339)
	}
	if user.LastLoginAt != nil {
		lastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET
			status = ?,
			display_name = ?,
			given_name = ?,
			family_name = ?,
			email = ?,
			phone = ?,
			email_verified = ?,
			phone_verified = ?,
			locale = ?,
			password_credential = ?,
			webauthn_credential = ?,
			mfa_credential = ?,
			password_locked = ?,
			webauthn_locked = ?,
			mfa_locked = ?,
			failed_login_attempts_password = ?,
			failed_login_attempts_webauthn = ?,
			failed_login_attempts_mfa = ?,
			roles = ?,
			groups = ?,
			attributes = ?,
			updated_at = ?,
			last_login_at = ?,
			federated_idp = ?,
			federated_id = ?,
			trusted_devices = ?
		WHERE id = ? AND tenant = ? AND realm = ?
	`,
		user.Status,
		user.DisplayName,
		user.GivenName,
		user.FamilyName,
		user.Email,
		user.Phone,
		user.EmailVerified,
		user.PhoneVerified,
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
		string(attributesJSON),
		updatedAt,
		lastLoginAt,
		user.FederatedIDP,
		user.FederatedID,
		string(trustedDevicesJSON),
		user.ID,
		user.Tenant,
		user.Realm,
	)
	return err
}

func (s *SQLiteUserDB) GetUserStats(ctx context.Context, tenant, realm string) (*model.UserStats, error) {
	var stats model.UserStats

	// Query to get all user statistics in a single query
	err := s.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_users,
			COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive_users,
			COUNT(CASE WHEN status = 'locked' THEN 1 END) as locked_users,
			COUNT(CASE WHEN email_verified = 1 THEN 1 END) as email_verified,
			COUNT(CASE WHEN phone_verified = 1 THEN 1 END) as phone_verified,
			COUNT(CASE WHEN webauthn_credential IS NOT NULL AND webauthn_credential != '' THEN 1 END) as webauthn_enabled,
			COUNT(CASE WHEN mfa_credential IS NOT NULL AND mfa_credential != '' THEN 1 END) as mfa_enabled,
			COUNT(CASE WHEN federated_idp IS NOT NULL THEN 1 END) as federated_users
		FROM users 
		WHERE tenant = ? AND realm = ?
	`, tenant, realm).Scan(
		&stats.TotalUsers,
		&stats.ActiveUsers,
		&stats.InactiveUsers,
		&stats.LockedUsers,
		&stats.EmailVerified,
		&stats.PhoneVerified,
		&stats.WebAuthnEnabled,
		&stats.MFAEnabled,
		&stats.FederatedUsers,
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

func (s *SQLiteUserDB) ListUsersWithPagination(ctx context.Context, tenant, realm string, offset, limit int) ([]model.User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant, realm, username,
		       status,
		       display_name, given_name, family_name,
		       email, phone, email_verified, phone_verified,
		       locale,
		       password_credential, webauthn_credential, mfa_credential,
		       password_locked, webauthn_locked, mfa_locked,
		       failed_login_attempts_password, failed_login_attempts_webauthn, failed_login_attempts_mfa,
		       roles, groups, attributes,
		       created_at, updated_at, last_login_at,
		       federated_idp, federated_id,
		       trusted_devices
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
		var user model.User
		var rolesJSON, groupsJSON, attributesJSON, trustedDevicesJSON string
		var createdAt, updatedAt string
		var lastLoginAt sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Tenant,
			&user.Realm,
			&user.Username,
			&user.Status,
			&user.DisplayName,
			&user.GivenName,
			&user.FamilyName,
			&user.Email,
			&user.Phone,
			&user.EmailVerified,
			&user.PhoneVerified,
			&user.Locale,
			&user.PasswordCredential,
			&user.WebAuthnCredential,
			&user.MFACredential,
			&user.PasswordLocked,
			&user.WebAuthnLocked,
			&user.MFALocked,
			&user.FailedLoginAttemptsPassword,
			&user.FailedLoginAttemptsWebAuthn,
			&user.FailedLoginAttemptsMFA,
			&rolesJSON,
			&groupsJSON,
			&attributesJSON,
			&createdAt,
			&updatedAt,
			&lastLoginAt,
			&user.FederatedIDP,
			&user.FederatedID,
			&trustedDevicesJSON,
		)
		if err != nil {
			return nil, err
		}

		// Parse timestamps
		user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		if lastLoginAt.Valid {
			lastLogin, _ := time.Parse(time.RFC3339, lastLoginAt.String)
			user.LastLoginAt = &lastLogin
		}

		// Parse JSON fields
		user.Roles = []string{}
		user.Groups = []string{}
		user.Attributes = map[string]string{}

		if rolesJSON != "" && rolesJSON != "null" {
			_ = json.Unmarshal([]byte(rolesJSON), &user.Roles)
		}
		if groupsJSON != "" && groupsJSON != "null" {
			_ = json.Unmarshal([]byte(groupsJSON), &user.Groups)
		}
		if attributesJSON != "" && attributesJSON != "null" {
			_ = json.Unmarshal([]byte(attributesJSON), &user.Attributes)
		}

		user.TrustedDevices = trustedDevicesJSON
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// DeleteUser deletes a user by username
func (s *SQLiteUserDB) DeleteUser(ctx context.Context, tenant, realm, username string) error {
	query := `
		DELETE FROM users 
		WHERE tenant = ? AND realm = ? AND username = ?
	`

	_, err := s.db.ExecContext(ctx, query, tenant, realm, username)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// No error if user doesn't exist (idempotent)
	return nil
}
