package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"goiam/internal/db/model"
	"time"

	"github.com/google/uuid"
)

func CreateUser(ctx context.Context, user model.User) error {
	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Convert roles, groups, attributes to JSON strings
	rolesJSON, _ := json.Marshal(user.Roles)
	groupsJSON, _ := json.Marshal(user.Groups)
	attributesJSON, _ := json.Marshal(user.Attributes)
	trustedJSON, _ := json.Marshal(user.TrustedDevices)

	_, err := DB.ExecContext(ctx, `
		INSERT INTO users (
			id, username, password_hash,
			display_name, given_name, family_name,
			email, phone,
			email_verified, phone_verified,
			roles, groups, attributes,
			created_at, updated_at,
			last_login_at,
			federated_idp, federated_id,
			trusted_devices,
			failed_login_attempts, last_failed_login_at, account_locked
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		user.ID,
		user.Username,
		user.PasswordHash,
		user.DisplayName,
		user.GivenName,
		user.FamilyName,
		user.Email,
		user.Phone,
		user.EmailVerified,
		user.PhoneVerified,
		string(rolesJSON),
		string(groupsJSON),
		string(attributesJSON),
		user.CreatedAt,
		user.UpdatedAt,
		user.LastLoginAt,
		user.FederatedIDP,
		user.FederatedID,
		string(trustedJSON),
		user.FailedLoginAttempts,
		user.LastFailedLoginAt,
		user.AccountLocked,
	)

	if err != nil {
		return err
	}
	return nil
}

func GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	row := DB.QueryRowContext(ctx, `
		SELECT id, username, password_hash,
		       display_name, given_name, family_name,
		       email, phone,
		       email_verified, phone_verified,
		       roles, groups, attributes,
		       created_at, updated_at, last_login_at,
		       federated_idp, federated_id,
		       trusted_devices,
		       failed_login_attempts, last_failed_login_at, account_locked
		FROM users WHERE username = ?
	`, username)

	var user model.User
	var rolesJSON, groupsJSON, attributesJSON, trustedJSON string

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.DisplayName,
		&user.GivenName,
		&user.FamilyName,
		&user.Email,
		&user.Phone,
		&user.EmailVerified,
		&user.PhoneVerified,
		&rolesJSON,
		&groupsJSON,
		&attributesJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
		&user.FederatedIDP,
		&user.FederatedID,
		&trustedJSON,
		&user.FailedLoginAttempts,
		&user.LastFailedLoginAt,
		&user.AccountLocked,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	_ = json.Unmarshal([]byte(rolesJSON), &user.Roles)
	_ = json.Unmarshal([]byte(groupsJSON), &user.Groups)
	_ = json.Unmarshal([]byte(attributesJSON), &user.Attributes)
	_ = json.Unmarshal([]byte(trustedJSON), &user.TrustedDevices)

	return &user, nil
}

func UpdateUser(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()

	rolesJSON, _ := json.Marshal(user.Roles)
	groupsJSON, _ := json.Marshal(user.Groups)
	attributesJSON, _ := json.Marshal(user.Attributes)
	trustedJSON, _ := json.Marshal(user.TrustedDevices)

	_, err := DB.ExecContext(ctx, `
		UPDATE users SET
			display_name = ?,
			given_name = ?,
			family_name = ?,
			email = ?,
			phone = ?,
			email_verified = ?,
			phone_verified = ?,
			roles = ?,
			groups = ?,
			attributes = ?,
			updated_at = ?,
			last_login_at = ?,
			federated_idp = ?,
			federated_id = ?,
			trusted_devices = ?,
			failed_login_attempts = ?,
			last_failed_login_at = ?,
			account_locked = ?
		WHERE id = ?
	`,
		user.DisplayName,
		user.GivenName,
		user.FamilyName,
		user.Email,
		user.Phone,
		user.EmailVerified,
		user.PhoneVerified,
		string(rolesJSON),
		string(groupsJSON),
		string(attributesJSON),
		user.UpdatedAt,
		user.LastLoginAt,
		user.FederatedIDP,
		user.FederatedID,
		string(trustedJSON),
		user.FailedLoginAttempts,
		user.LastFailedLoginAt,
		user.AccountLocked,
		user.ID,
	)
	return err
}
