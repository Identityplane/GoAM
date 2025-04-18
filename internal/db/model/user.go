package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Interface for the user db
type UserDB interface {
	CreateUser(ctx context.Context, user User) error
	GetUserByUsername(ctx context.Context, tenant, realm, username string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	ListUsers(ctx context.Context, tenant, realm string) ([]User, error)
	ListUsersWithPagination(ctx context.Context, tenant, realm string, offset, limit int) ([]User, error)
	CountUsers(ctx context.Context, tenant, realm string) (int64, error)
	GetUserStats(ctx context.Context, tenant, realm string) (*UserStats, error)
	DeleteUser(ctx context.Context, tenant, realm, username string) error
}

// User represents a user in the system
// @description User information and attributes
type User struct {
	// Unique UUID for the user
	ID string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Organization Context
	Tenant   string `json:"tenant" example:"acme"`
	Realm    string `json:"realm" example:"customers"`
	Username string `json:"username" example:"john.doe"`

	// User status
	Status string `json:"status" example:"active"`

	// Identity Information
	DisplayName string `json:"display_name" example:"John Doe"`
	GivenName   string `json:"given_name" example:"John"`
	FamilyName  string `json:"family_name" example:"Doe"`

	// Additional contact information
	Email         string `json:"email" example:"john.doe@example.com"`
	Phone         string `json:"phone" example:"+1234567890"`
	EmailVerified bool   `json:"email_verified" example:"true"`
	PhoneVerified bool   `json:"phone_verified" example:"false"`

	// Locale
	Locale string `json:"locale" example:"en-US"`

	// Authentication credentials
	PasswordCredential string `json:"-"`
	WebAuthnCredential string `json:"-"`
	MFACredential      string `json:"-"`

	PasswordLocked bool `json:"password_locked" example:"false"`
	WebAuthnLocked bool `json:"webauthn_locked" example:"false"`
	MFALocked      bool `json:"mfa_locked" example:"false"`

	FailedLoginAttemptsPassword int `json:"failed_login_attempts_password" example:"0"`
	FailedLoginAttemptsWebAuthn int `json:"failed_login_attempts_webauthn" example:"0"`
	FailedLoginAttemptsMFA      int `json:"failed_login_attempts_mfa" example:"0"`

	// User roles and groups
	Roles  []string `json:"roles" example:"['admin', 'user']"`
	Groups []string `json:"groups" example:"['developers', 'support']"`

	// Extensibility
	Attributes map[string]string `json:"attributes" example:"{'department': 'IT', 'location': 'HQ'}"`

	// Audit
	CreatedAt   time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" example:"2024-01-01T00:00:00Z"`

	// Federation
	FederatedIDP *string `json:"federated_idp,omitempty" example:"google"`
	FederatedID  *string `json:"federated_id,omitempty" example:"123456789"`

	// Devices
	TrustedDevices string `json:"trusted_devices,omitempty" example:"device1,device2"`
}

// UserStats represents user statistics
// @description User statistics for a realm
type UserStats struct {
	TotalUsers      int64 `json:"total_users" example:"100"`
	ActiveUsers     int64 `json:"active_users" example:"80"`
	InactiveUsers   int64 `json:"inactive_users" example:"15"`
	LockedUsers     int64 `json:"locked_users" example:"5"`
	EmailVerified   int64 `json:"email_verified" example:"90"`
	PhoneVerified   int64 `json:"phone_verified" example:"70"`
	WebAuthnEnabled int64 `json:"webauthn_enabled" example:"30"`
	MFAEnabled      int64 `json:"mfa_enabled" example:"40"`
	FederatedUsers  int64 `json:"federated_users" example:"20"`
}

func (u *User) UnmarshalJSON(data []byte) error {
	// Make a mirror struct where time fields are strings
	type Alias User
	aux := &struct {
		CreatedAt   *string `json:"created_at"`
		UpdatedAt   *string `json:"updated_at"`
		LastLoginAt *string `json:"last_login_at"`
		*Alias
	}{
		Alias: (*Alias)(u),
	}

	// First, unmarshal into aux
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Now manually parse each timestamp
	parseTime := func(s *string) (*time.Time, error) {
		if s == nil || *s == "" {
			return nil, nil
		}
		t, err := time.Parse(time.RFC3339, *s)
		if err != nil {
			return nil, err
		}
		return &t, nil
	}

	var err error
	if t, err := parseTime(aux.CreatedAt); err != nil {
		return fmt.Errorf("invalid created_at: %w", err)
	} else if t != nil {
		u.CreatedAt = *t
	}
	if t, err := parseTime(aux.UpdatedAt); err != nil {
		return fmt.Errorf("invalid updated_at: %w", err)
	} else if t != nil {
		u.UpdatedAt = *t
	}
	if u.LastLoginAt, err = parseTime(aux.LastLoginAt); err != nil {
		return fmt.Errorf("invalid last_login_at: %w", err)
	}

	return nil
}
