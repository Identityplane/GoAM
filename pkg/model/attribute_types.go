package model

import "time"

const (
	AttributeTypeTOTP = "identityplane:totp"
)

// TOTPAttributeValue is the attribute value for TOTP
// @description TOTP information
type TOTPAttributeValue struct {

	// @description The secret key for the TOTP
	SecretKey string `json:"secret" example:"1234567890"`

	// @description Whether the TOTP is locked
	Locked bool `json:"locked" example:"false"`

	// @description The number of failed attempts
	FailedAttempts int `json:"failed_attempts" example:"0"`
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
