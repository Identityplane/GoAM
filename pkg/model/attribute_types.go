package model

import (
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

const (
	AttributeTypeTOTP         = "identityplane:totp"
	AttributeTypeUsername     = "identityplane:username"
	AttributeTypeGitHub       = "identityplane:github"
	AttributeTypeTelegram     = "identityplane:telegram"
	AttributeTypePassword     = "identityplane:password"
	AttributeTypeEmail        = "identityplane:email"
	AttributeTypePhone        = "identityplane:phone"
	AttributeTypePasskey      = "identityplane:passkey"
	AttributeTypeEntitlements = "identityplane:entitlements"
	AttributeTypeYubico       = "identityplane:yubico"
	AttributeTypeDevice       = "identityplane:device"
)

// CredentialAttributeValue is the attribute value for credentials such as password otp etc
// Credentials can be lockled and we track the number of failed attempts
type CredentialAttributeValue struct {

	// @description Whether the credential is locked
	Locked bool `json:"locked" example:"false"`

	// @description The number of failed attempts
	FailedAttempts int `json:"failed_attempts" example:"0"`
}

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

// UsernameAttributeValue is the attribute value for usernames
// @description Username information
type UsernameAttributeValue struct {
	Username string `json:"username" example:"john.doe"`
}

// GitHubAttributeValue is the attribute value for GitHub
// @description GitHub information
type GitHubAttributeValue struct {
	GitHubUserID       string `json:"github_user_id" example:"1234567890"`
	GitHubRefreshToken string `json:"github_refresh_token" example:"1234567890"`
	GitHubEmail        string `json:"github_email" example:"john.doe@example.com"`
	GitHubAvatarURL    string `json:"github_avatar_url" example:"https://example.com/avatar.jpg"`
	GitHubUsername     string `json:"github_username" example:"john.doe"`
	GitHubAccessToken  string `json:"github_access_token" example:"1234567890"`
	GitHubTokenType    string `json:"github_token_type" example:"bearer"`
	GitHubScope        string `json:"github_scope" example:"user:email"`
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
	PasswordHash         string     `json:"password_hash" example:"password"`
	Locked               bool       `json:"locked" example:"false"`
	FailedAttempts       int        `json:"failed_attempts" example:"0"`
	LastCorrectTimestamp *time.Time `json:"last_correct_timestamp,omitempty" example:"2024-01-01T00:00:00Z"`
}

// EmailAttributeValue is the attribute value for emails
// @description Email information
type EmailAttributeValue struct {
	Email      string     `json:"email" example:"john.doe@example.com"`
	Verified   bool       `json:"verified" example:"true"`
	VerifiedAt *time.Time `json:"verified_at" example:"2024-01-01T00:00:00Z"`

	OtpFailedAttempts int  `json:"otp_failed_attempts" example:"0"`
	OtpLocked         bool `json:"otp_locked" example:"false"`
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
	RPID               string               `json:"rp_id" example:"example.com"`
	CredentialID       string               `json:"credential_id" example:"1234567890"`
	DisplayName        string               `json:"display_name" example:"John Doe"`
	WebAuthnCredential *webauthn.Credential `json:"webauthn_credential" example:"{}"`
	LastUsedAt         *time.Time           `json:"last_used_at" example:"2024-01-01T00:00:00Z"`
}

// UserProfileAttributeValue is the attribute value for user profiles
// @description User profile information
type UserProfileAttributeValue struct {
	DisplayName string `json:"display_name" example:"John Doe"`
	GivenName   string `json:"given_name" example:"John"`
	FamilyName  string `json:"family_name" example:"Doe"`
	Locale      string `json:"locale" example:"en-US"`
	PictureUri  string `json:"picture_uri" example:"https://example.com/profile.jpg"`
}

// TelegramAttributeValue is the attribute value for Telegram accounts
// @description Telegram information
type TelegramAttributeValue struct {
	TelegramUserID    string `json:"telegram_user_id" example:"1234567890"`
	TelegramUsername  string `json:"telegram_username" example:"johndoe"`
	TelegramFirstName string `json:"telegram_first_name" example:"John"`
	TelegramPhotoURL  string `json:"telegram_photo_url" example:"https://t.me/i/userpic/123/photo.jpg"`
	TelegramAuthDate  int64  `json:"telegram_auth_date" example:"1753278987"`
}

// DeviceAttributeValue is the attribute value for devices
// @description Device information
type DeviceAttributeValue struct {
	DeviceID         string `json:"device_id" example:"1234567890"`
	DeviceSecretHash string `json:"device_secret_hash" example:"1234567890"`
	DeviceName       string `json:"device_name" example:"John Doe's iPhone"`
	DeviceType       string `json:"device_type" example:"mobile"`
	DeviceOS         string `json:"device_os" example:"iOS"`
	DeviceOSVersion  string `json:"device_os_version" example:"15.0"`
	DeviceModel      string `json:"device_model" example:"iPhone 12"`
	DeviceIP         string `json:"device_ip" example:"192.168.1.100"`
	DeviceUserAgent  string `json:"device_user_agent" example:"Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1"`
	CookieName       string `json:"cookie_name" example:"session"`

	SessionDuration     int        `json:"session_duration" example:"3600"`
	SessionExpiry       *time.Time `json:"session_expiry" example:"2024-01-01T00:00:00Z"`
	SessionRefreshAfter int        `json:"session_refresh_after" example:"1800"`
	SessionFirstLogin   *time.Time `json:"session_first_login" example:"2024-01-01T00:00:00Z"`
	SessionLastActivity *time.Time `json:"session_last_activity" example:"2024-01-01T00:00:00Z"`
}

// EntitlementSet is a set of entitlements
// @description Entitlement set information
type EntitlementSetAttributeValue struct {
	Entitlements []Entitlement `json:"entitlements" example:"['admin', 'user']"`
}

type Entitlement struct {
	Description string     `json:"description" example:"Admin"`
	Resource    string     `json:"resource" example:"arn:identityplane:acme:customers:users:123"`
	Action      string     `json:"action" example:"read"`
	Effect      EffectType `json:"effect" example:"allow"`
}

type EffectType string

const (
	EffectTypeAllow EffectType = "allow"
	EffectTypeDeny  EffectType = "deny"
)

type YubicoAttributeValue struct {

	// @description The public id for the yubikey
	PublicID string `json:"public_id" example:"vvcijgklnrbf"`

	CredentialAttributeValue
}
