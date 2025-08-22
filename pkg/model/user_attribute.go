package model

import "time"

// User represents a aatribute linked to a user
// Attributes can be:
// - credentials such as password, passkey, links to socials accounts
// - PII such as user profile picture
// - Application specific attributes such as consent, settings, etc.
// - Entitlements such as roles, groups, etc.
// - Business specific attributes such as KYC status, etc.
// @description User information and attributes
type UserAttribute struct {
	ID string `json:"id" db:"id" example:"123e4567-e89b-12d3-a456-426614174000"` // The id of the attribute

	UserID string `json:"user_id" db:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"` // The id of the user who this attribute belongs to
	Tenant string `json:"tenant" db:"tenant" example:"acme"`                                   // The tenant of the user who this attribute belongs to
	Realm  string `json:"realm" db:"realm" example:"customers"`                                // The realm of the user who this attribute belongs to

	Index string `json:"index" db:"index_value" example:""` // the index can be used to lookup a user by attribute, e.g. by social idp login name. Index should be unique per realm
	Type  string `json:"type" db:"type" example:"password"` // the type of the attribute, e.g. password, email, phone, etc.
	Value any    `json:"value" db:"value"`                  // the value of the attribute, e.g. password, email, phone, etc.

	CreatedAt time.Time `json:"created_at" db:"created_at" example:"2024-01-01T00:00:00Z"` // The time the attribute was created
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" example:"2024-01-01T00:00:00Z"` // The time the attribute was last updated
}

// SocialAttributeValue is the attribute value for social accounts
// @description Social account information
type SocialAttributeValue struct {
	SocialIDP string `json:"social_idp" example:"google"`
	SocialID  string `json:"social_id" example:"1234567890"`
}

// PasswordAttributeValue is the attribute value for passwords
// @description Password information
type PasswordAttributeValue struct {
	PasswordHash   string `json:"password_hash" example:"password"`
	Locked         bool   `json:"locked" example:"false"`
	FailedAttempts int    `json:"failed_attempts" example:"0"`
}

// EmailAttributeValue is the attribute value for emails
// @description Email information
type EmailAttributeValue struct {
	Email      string     `json:"email" example:"john.doe@example.com"`
	Verified   bool       `json:"verified" example:"true"`
	VerifiedAt *time.Time `json:"verified_at" example:"2024-01-01T00:00:00Z"`
}

// PhoneAttributeValue is the attribute value for phones
// @description Phone information
type PhoneAttributeValue struct {
	Phone      string     `json:"phone" example:"+1234567890"`
	Verified   bool       `json:"verified" example:"true"`
	VerifiedAt *time.Time `json:"verified_at" example:"2024-01-01T00:00:00Z"`
}

// TOTPAttributeValue is the attribute value for TOTP
// @description TOTP information
type TOTPAttributeValue struct {
	SecretKey      string `json:"secret" example:"1234567890"`
	Locked         bool   `json:"locked" example:"false"`
	FailedAttempts int    `json:"failed_attempts" example:"0"`
}

// PasskeyAttributeValue is the attribute value for passkeys
// @description Passkey information
type PasskeyAttributeValue struct {
	RPID               string `json:"rp_id" example:"example.com"`
	WebAuthnCredential string `json:"webauthn_credential" example:"{}"`
}

// UserProfileAttributeValue is the attribute value for user profiles
// @description User profile information
type UserProfileAttributeValue struct {
	DisplayName string `json:"display_name" example:"John Doe"`
	GivenName   string `json:"given_name" example:"John"`
	FamilyName  string `json:"family_name" example:"Doe"`
}

// UserPictureAttributeValue is the attribute value for user pictures
// @description User picture information
type UserPictureAttributeValue struct {
	Url string `json:"picture" example:"https://example.com/profile.jpg"`
}

// DeviceAttributeValue is the attribute value for devices
// @description Device information
type DeviceAttributeValue struct {
	DeviceID         string `json:"device_id" example:"1234567890"`
	DeviceSecretHash string `json:"device_secret_hash" example:"1234567890"`

	DeviceName      string `json:"device_name" example:"John Doe's iPhone"`
	DeviceType      string `json:"device_type" example:"mobile"`
	DeviceOS        string `json:"device_os" example:"iOS"`
	DeviceOSVersion string `json:"device_os_version" example:"15.0"`
	DeviceModel     string `json:"device_model" example:"iPhone 12"`
	DeviceIP        string `json:"device_ip" example:"192.168.1.100"`
}
