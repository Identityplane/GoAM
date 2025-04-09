package model

import (
	"time"
)

type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	PasswordHash  string `json:"-"`
	DisplayName   string `json:"display_name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	EmailVerified bool   `json:"email_verified"`
	PhoneVerified bool   `json:"phone_verified"`

	Roles      []string          `json:"roles"`
	Groups     []string          `json:"groups"`
	Attributes map[string]string `json:"attributes"`

	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`

	// Federation
	FederatedIDP *string `json:"federated_idp,omitempty"`
	FederatedID  *string `json:"federated_id,omitempty"`

	// Device trust
	TrustedDevices []string `json:"trusted_devices,omitempty"`

	// 🔐 Login security
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	LastFailedLoginAt   *time.Time `json:"last_failed_login_at,omitempty"`
	AccountLocked       bool       `json:"account_locked"`
}
