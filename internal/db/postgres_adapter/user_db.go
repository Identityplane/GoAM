package postgres_adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"goiam/internal/logger"
	"goiam/internal/model"
	"time"

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

		logger.DebugNoContext("Warning: failed to check if users table exists: %v", err)
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

	// Convert boolean fields to PostgreSQL boolean
	emailVerified := user.EmailVerified
	phoneVerified := user.PhoneVerified
	passwordLocked := user.PasswordLocked
	webauthnLocked := user.WebAuthnLocked
	mfaLocked := user.MFALocked

	_, err := p.db.Exec(ctx, `
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
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35)
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
		emailVerified,
		phoneVerified,
		user.LoginIdentifier,
		user.Locale,
		user.PasswordCredential,
		user.WebAuthnCredential,
		user.MFACredential,
		passwordLocked,
		webauthnLocked,
		mfaLocked,
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

func (p *PostgresUserDB) GetUserByUsername(ctx context.Context, tenant, realm, username string) (*model.User, error) {
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
		WHERE tenant = $1 AND realm = $2 AND username = $3
	`, tenant, realm, username)

	return p.scanUserFromRow(row)
}

func (p *PostgresUserDB) UpdateUser(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()

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

	// Convert boolean fields to PostgreSQL boolean
	emailVerified := user.EmailVerified
	phoneVerified := user.PhoneVerified
	passwordLocked := user.PasswordLocked
	webauthnLocked := user.WebAuthnLocked
	mfaLocked := user.MFALocked

	_, err := p.db.Exec(ctx, `
		UPDATE users SET
			status = $1,
			display_name = $2,
			given_name = $3,
			family_name = $4,
			profile_picture_uri = $5,
			email = $6,
			phone = $7,
			email_verified = $8,
			phone_verified = $9,
			login_identifier = $10,
			locale = $11,
			password_credential = $12,
			webauthn_credential = $13,
			mfa_credential = $14,
			password_locked = $15,
			webauthn_locked = $16,
			mfa_locked = $17,
			failed_login_attempts_password = $18,
			failed_login_attempts_webauthn = $19,
			failed_login_attempts_mfa = $20,
			roles = $21,
			groups = $22,
			entitlements = $23,
			consent = $24,
			attributes = $25,
			updated_at = $26,
			last_login_at = $27,
			federated_idp = $28,
			federated_id = $29,
			trusted_devices = $30
		WHERE id = $31 AND tenant = $32 AND realm = $33
	`,
		user.Status,
		user.DisplayName,
		user.GivenName,
		user.FamilyName,
		user.ProfilePictureURI,
		user.Email,
		user.Phone,
		emailVerified,
		phoneVerified,
		user.LoginIdentifier,
		user.Locale,
		user.PasswordCredential,
		user.WebAuthnCredential,
		user.MFACredential,
		passwordLocked,
		webauthnLocked,
		mfaLocked,
		user.FailedLoginAttemptsPassword,
		user.FailedLoginAttemptsWebAuthn,
		user.FailedLoginAttemptsMFA,
		string(rolesJSON),
		string(groupsJSON),
		string(entitlementsJSON),
		string(consentJSON),
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

func (p *PostgresUserDB) ListUsers(ctx context.Context, tenant, realm string) ([]model.User, error) {
	rows, err := p.db.Query(ctx, `
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
		WHERE tenant = $1 AND realm = $2
	`, tenant, realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		var rolesJSON, groupsJSON, attributesJSON, trustedDevicesJSON string
		var createdAt, updatedAt string
		var lastLoginAt *string
		var emailVerified, phoneVerified, passwordLocked, webauthnLocked, mfaLocked bool

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
			&emailVerified,
			&phoneVerified,
			&user.Locale,
			&user.PasswordCredential,
			&user.WebAuthnCredential,
			&user.MFACredential,
			&passwordLocked,
			&webauthnLocked,
			&mfaLocked,
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

		// Set boolean fields
		user.EmailVerified = emailVerified
		user.PhoneVerified = phoneVerified
		user.PasswordLocked = passwordLocked
		user.WebAuthnLocked = webauthnLocked
		user.MFALocked = mfaLocked

		// Parse timestamps
		user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		if lastLoginAt != nil {
			lastLogin, _ := time.Parse(time.RFC3339, *lastLoginAt)
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

func (p *PostgresUserDB) ListUsersWithPagination(ctx context.Context, tenant, realm string, offset, limit int) ([]model.User, error) {
	rows, err := p.db.Query(ctx, `
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
		WHERE tenant = $1 AND realm = $2
		ORDER BY username
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
		var rolesStr, groupsStr, entitlementsStr, consentStr, attributesStr, trustedDevicesStr string

		err := rows.Scan(
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
			&rolesStr,
			&groupsStr,
			&entitlementsStr,
			&consentStr,
			&attributesStr,
			&createdAt,
			&updatedAt,
			&lastLoginAt,
			&user.FederatedIDP,
			&user.FederatedID,
			&trustedDevicesStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		// Set time fields
		user.CreatedAt = createdAt
		user.UpdatedAt = updatedAt
		user.LastLoginAt = lastLoginAt

		// Parse JSON fields
		if err := json.Unmarshal([]byte(rolesStr), &user.Roles); err != nil {
			return nil, fmt.Errorf("failed to unmarshal roles: %w", err)
		}
		if err := json.Unmarshal([]byte(groupsStr), &user.Groups); err != nil {
			return nil, fmt.Errorf("failed to unmarshal groups: %w", err)
		}
		if err := json.Unmarshal([]byte(entitlementsStr), &user.Entitlements); err != nil {
			return nil, fmt.Errorf("failed to unmarshal entitlements: %w", err)
		}
		if err := json.Unmarshal([]byte(consentStr), &user.Consent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal consent: %w", err)
		}
		if err := json.Unmarshal([]byte(attributesStr), &user.Attributes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal attributes: %w", err)
		}
		if err := json.Unmarshal([]byte(trustedDevicesStr), &user.TrustedDevices); err != nil {
			return nil, fmt.Errorf("failed to unmarshal trusted devices: %w", err)
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

	// Query to get all user statistics in a single query
	err := p.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_users,
			COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive_users,
			COUNT(CASE WHEN status = 'locked' THEN 1 END) as locked_users,
			COUNT(CASE WHEN email_verified = true THEN 1 END) as email_verified,
			COUNT(CASE WHEN phone_verified = true THEN 1 END) as phone_verified,
			COUNT(CASE WHEN webauthn_credential IS NOT NULL AND webauthn_credential != '' THEN 1 END) as webauthn_enabled,
			COUNT(CASE WHEN mfa_credential IS NOT NULL AND mfa_credential != '' THEN 1 END) as mfa_enabled,
			COUNT(CASE WHEN federated_idp IS NOT NULL THEN 1 END) as federated_users
		FROM users 
		WHERE tenant = $1 AND realm = $2
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

// DeleteUser deletes a user by username
func (p *PostgresUserDB) DeleteUser(ctx context.Context, tenant, realm, username string) error {
	query := `
		DELETE FROM users 
		WHERE tenant = $1 AND realm = $2 AND username = $3
	`

	_, err := p.db.Exec(ctx, query, tenant, realm, username)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// No error if user doesn't exist (idempotent)
	return nil
}

// scanUserFromRow scans a user from a database row
func (p *PostgresUserDB) scanUserFromRow(row pgx.Row) (*model.User, error) {
	var user model.User
	var rolesJSON, groupsJSON, attributesJSON, trustedDevicesJSON, entitlementsJSON, consentJSON string
	var createdAt, updatedAt time.Time
	var lastLoginAt *time.Time
	var emailVerified, phoneVerified, passwordLocked, webauthnLocked, mfaLocked bool

	err := row.Scan(
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
		&emailVerified,
		&phoneVerified,
		&user.LoginIdentifier,
		&user.Locale,
		&user.PasswordCredential,
		&user.WebAuthnCredential,
		&user.MFACredential,
		&passwordLocked,
		&webauthnLocked,
		&mfaLocked,
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
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	// Set boolean fields
	user.EmailVerified = emailVerified
	user.PhoneVerified = phoneVerified
	user.PasswordLocked = passwordLocked
	user.WebAuthnLocked = webauthnLocked
	user.MFALocked = mfaLocked

	// Set timestamps
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
	if lastLoginAt != nil {
		user.LastLoginAt = lastLoginAt
	}

	// Parse JSON fields
	user.Roles = []string{}
	user.Groups = []string{}
	user.Attributes = map[string]string{}
	user.Entitlements = []string{}
	user.Consent = []string{}

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

	user.TrustedDevices = trustedDevicesJSON

	return &user, nil
}

func (p *PostgresUserDB) GetUserByID(ctx context.Context, tenant, realm, userID string) (*model.User, error) {
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

	return p.scanUserFromRow(row)
}

func (p *PostgresUserDB) GetUserByEmail(ctx context.Context, tenant, realm, email string) (*model.User, error) {
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
		WHERE tenant = $1 AND realm = $2 AND email = $3
	`, tenant, realm, email)

	return p.scanUserFromRow(row)
}

func (p *PostgresUserDB) GetUserByLoginIdentifier(ctx context.Context, tenant, realm, loginIdentifier string) (*model.User, error) {
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
		WHERE tenant = $1 AND realm = $2 AND login_identifier = $3
	`, tenant, realm, loginIdentifier)

	return p.scanUserFromRow(row)
}
