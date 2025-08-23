package model

import (
	"encoding/json"
	"fmt"
	"time"
)

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

	// Profile Information
	ProfilePictureURI string `json:"profile_picture_uri" example:"https://example.com/profile.jpg"`

	// Additional contact information
	Email         string `json:"email" example:"john.doe@example.com"`
	Phone         string `json:"phone" example:"+1234567890"`
	EmailVerified bool   `json:"email_verified" example:"true"`
	PhoneVerified bool   `json:"phone_verified" example:"false"`

	// Login Information
	LoginIdentifier string `json:"login_identifier" example:"john.doe@example.com"`

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

	// User roles, groups and entitlements
	Roles        []string `json:"roles" example:"['admin', 'user']"`
	Groups       []string `json:"groups" example:"['developers', 'support']"`
	Entitlements []string `json:"entitlements" example:"['read:users', 'write:users']"`

	// User consents
	Consent []string `json:"consent,omitempty" example:"['marketing', 'analytics']"`

	// Extensibility
	Attributes map[string]string `json:"attributes,omitempty" example:"{'key1': 'value1', 'key2': 'value2'}"`

	// Audit
	CreatedAt   time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" example:"2024-01-01T00:00:00Z"`

	// Federation
	FederatedIDP *string `json:"federated_idp,omitempty" example:"google"`
	FederatedID  *string `json:"federated_id,omitempty" example:"123456789"`

	// Devices
	TrustedDevices []string `json:"trusted_devices,omitempty" example:"['device1', 'device2']"`

	// Attributes
	UserAttributes []UserAttribute `json:"user_attributes,omitempty"`
}

// GetAttributesByType returns all attributes of a specific type for the user
func (u *User) GetAttributesByType(attrType string) []UserAttribute {
	var filtered []UserAttribute
	for _, attr := range u.UserAttributes {
		if attr.Type == attrType {
			filtered = append(filtered, attr)
		}
	}
	return filtered
}

// GetAttribute returns the first attribute of the specified type and converts its value to the target type
// Returns an error if multiple attributes of the same type exist, nil if no attribute exists, or an error if conversion fails
func GetAttribute[T any](u *User, attrType string) (*T, *UserAttribute, error) {
	attrs := u.GetAttributesByType(attrType)
	if len(attrs) == 0 {
		return nil, nil, nil
	}

	if len(attrs) > 1 {
		return nil, nil, fmt.Errorf("multiple attributes of type '%s' found, use GetAttributesByType instead", attrType)
	}

	// Get the single attribute of this type
	attr := attrs[0]

	// Try to convert the value to the target type
	if converted, ok := attr.Value.(T); ok {
		return &converted, &attr, nil
	}

	// If direct conversion fails, try to convert from map[string]interface{} (for database stored values)
	if mapValue, ok := attr.Value.(map[string]interface{}); ok {
		// Convert map to JSON and then to the target type
		jsonData, err := json.Marshal(mapValue)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal attribute value to JSON: %w", err)
		}

		var result T
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal attribute value to %T: %w", result, err)
		}

		return &result, &attr, nil
	}

	return nil, nil, fmt.Errorf("failed to convert attribute value from %T to %T", attr.Value, *new(T))
}

// GetAttributes returns all attributes of the specified type and converts their values to the target type.
// It always returns a slice of attributes, even if there are no attributes of the specified type.
// Only if there is an attribute that cannot be converted to the target type, an error is returned.
func GetAttributes[T any](u *User, attrType string) ([]T, []*UserAttribute, error) {
	attrs := u.GetAttributesByType(attrType)
	if len(attrs) == 0 {
		return []T{}, []*UserAttribute{}, nil
	}

	var result []T
	var attributes []*UserAttribute

	for _, attr := range attrs {
		// Try to convert the value to the target type
		if converted, ok := attr.Value.(T); ok {
			result = append(result, converted)
			attributes = append(attributes, &attr)
			continue
		}

		// If direct conversion fails, try to convert from map[string]interface{} (for database stored values)
		if mapValue, ok := attr.Value.(map[string]interface{}); ok {
			// Convert map to JSON and then to the target type
			jsonData, err := json.Marshal(mapValue)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal attribute value to JSON: %w", err)
			}

			var converted T
			if err := json.Unmarshal(jsonData, &converted); err != nil {
				return nil, nil, fmt.Errorf("failed to unmarshal attribute value to %T: %w", converted, err)
			}

			result = append(result, converted)
			attributes = append(attributes, &attr)
			continue
		}

		// If conversion fails, return error
		return nil, nil, fmt.Errorf("failed to convert attribute value from %T to %T", attr.Value, *new(T))
	}

	return result, attributes, nil
}

func (u *User) AddAttribute(attr *UserAttribute) {

	// Ensure the tenant, realm, and userid are set correctly
	attr.Tenant = u.Tenant
	attr.Realm = u.Realm
	attr.UserID = u.ID

	// Add the attribute to the user
	u.UserAttributes = append(u.UserAttributes, *attr)
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
