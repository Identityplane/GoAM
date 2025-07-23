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
		log := logger.GetLogger()
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

	// Convert JSON fields to strings
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

	_, err := s.db.ExecContext(ctx, `
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
		return err
	}
	return nil
}

func (s *SQLiteUserDB) GetUserByUsername(ctx context.Context, tenant, realm, username string) (*model.User, error) {
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
		WHERE tenant = ? AND realm = ? AND username = ?
	`, tenant, realm, username)

	return s.scanUserFromRow(row)
}

// scanUserFromRow scans a user from a database row
func (s *SQLiteUserDB) scanUserFromRow(scanner interface{}) (*model.User, error) {
	var user model.User
	var rolesJSON, groupsJSON, attributesJSON, trustedDevicesJSON, entitlementsJSON, consentJSON string
	var createdAt, updatedAt string
	var lastLoginAt sql.NullString

	var err error
	switch r := scanner.(type) {
	case *sql.Row:
		err = r.Scan(
			&user.ID,
			&user.Tenant,
			&user.Realm,
			&user.Username,
			&user.Status,
			&user.DisplayName,
			&user.GivenName,
			&user.FamilyName,
			&user.ProfilePictureURI,
			&user.Email,
			&user.Phone,
			&user.EmailVerified,
			&user.PhoneVerified,
			&user.LoginIdentifier,
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
			&entitlementsJSON,
			&consentJSON,
			&attributesJSON,
			&createdAt,
			&updatedAt,
			&lastLoginAt,
			&user.FederatedIDP,
			&user.FederatedID,
			&trustedDevicesJSON,
		)
	case *sql.Rows:
		err = r.Scan(
			&user.ID,
			&user.Tenant,
			&user.Realm,
			&user.Username,
			&user.Status,
			&user.DisplayName,
			&user.GivenName,
			&user.FamilyName,
			&user.ProfilePictureURI,
			&user.Email,
			&user.Phone,
			&user.EmailVerified,
			&user.PhoneVerified,
			&user.LoginIdentifier,
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
			&entitlementsJSON,
			&consentJSON,
			&attributesJSON,
			&createdAt,
			&updatedAt,
			&lastLoginAt,
			&user.FederatedIDP,
			&user.FederatedID,
			&trustedDevicesJSON,
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

	// Parse JSON fields
	user.Roles = []string{}
	user.Groups = []string{}
	user.Attributes = map[string]string{}
	user.Entitlements = []string{}
	user.Consent = []string{}
	user.TrustedDevices = []string{}

	if rolesJSON != "" && rolesJSON != "null" {
		_ = json.Unmarshal([]byte(rolesJSON), &user.Roles)
	}
	if groupsJSON != "" && groupsJSON != "null" {
		_ = json.Unmarshal([]byte(groupsJSON), &user.Groups)
	}
	if attributesJSON != "" && attributesJSON != "null" {
		_ = json.Unmarshal([]byte(attributesJSON), &user.Attributes)
	}
	if entitlementsJSON != "" && entitlementsJSON != "null" {
		_ = json.Unmarshal([]byte(entitlementsJSON), &user.Entitlements)
	}
	if consentJSON != "" && consentJSON != "null" {
		_ = json.Unmarshal([]byte(consentJSON), &user.Consent)
	}
	if trustedDevicesJSON != "" && trustedDevicesJSON != "null" {
		_ = json.Unmarshal([]byte(trustedDevicesJSON), &user.TrustedDevices)
	}

	return &user, nil
}

func (s *SQLiteUserDB) GetUserByID(ctx context.Context, tenant, realm, userID string) (*model.User, error) {
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

	return s.scanUserFromRow(row)
}

func (s *SQLiteUserDB) UpdateUser(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()

	// Convert JSON fields to strings
	rolesJSON, _ := json.Marshal(user.Roles)
	groupsJSON, _ := json.Marshal(user.Groups)
	attributesJSON, _ := json.Marshal(user.Attributes)
	trustedDevicesJSON, _ := json.Marshal(user.TrustedDevices)
	entitlementsJSON, _ := json.Marshal(user.Entitlements)
	consentJSON, _ := json.Marshal(user.Consent)

	// Handle time fields
	var lastLoginAt interface{}
	if user.LastLoginAt != nil {
		lastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET
			status = ?,
			display_name = ?,
			given_name = ?,
			family_name = ?,
			profile_picture_uri = ?,
			email = ?,
			phone = ?,
			email_verified = ?,
			phone_verified = ?,
			login_identifier = ?,
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
			entitlements = ?,
			consent = ?,
			attributes = ?,
			updated_at = ?,
			last_login_at = ?,
			federated_idp = ?,
			federated_id = ?,
			trusted_devices = ?
		WHERE tenant = ? AND realm = ? AND id = ?
	`,
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
		user.UpdatedAt.Format(time.RFC3339),
		lastLoginAt,
		user.FederatedIDP,
		user.FederatedID,
		string(trustedDevicesJSON),
		user.Tenant,
		user.Realm,
		user.ID,
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

// ListUsers implements model.UserDB.
func (s *SQLiteUserDB) ListUsers(ctx context.Context, tenant string, realm string) ([]model.User, error) {
	rows, err := s.db.QueryContext(ctx, `
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

func (s *SQLiteUserDB) GetUserByEmail(ctx context.Context, tenant, realm, email string) (*model.User, error) {
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
		WHERE tenant = ? AND realm = ? AND email = ?
	`, tenant, realm, email)

	return s.scanUserFromRow(row)
}

func (s *SQLiteUserDB) GetUserByLoginIdentifier(ctx context.Context, tenant, realm, loginIdentifier string) (*model.User, error) {
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
		WHERE tenant = ? AND realm = ? AND login_identifier = ?
	`, tenant, realm, loginIdentifier)

	return s.scanUserFromRow(row)
}

func (s *SQLiteUserDB) GetUserByFederatedIdentifier(ctx context.Context, tenant, realm, provider, identifier string) (*model.User, error) {
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
		WHERE tenant = ? AND realm = ? AND federated_idp = ? AND federated_id = ?
	`, tenant, realm, provider, identifier)

	return s.scanUserFromRow(row)
}
