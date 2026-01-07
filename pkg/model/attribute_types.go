package model

import (
	"github.com/Identityplane/GoAM/pkg/model/attributes"
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
	AttributeTypeOidc         = "identityplane:oidc"
)

// AttributeValue is the interface that all attribute value types must implement
// The GetIndex method returns a unique identifier for the attribute value
// that can be used for user lookup within a realm
// The IndexIsSensitive method returns whether the index should be omitted from JSON API responses
type AttributeValue interface {
	GetIndex() string
	IndexIsSensitive() bool
}

// Type aliases for backward compatibility - these point to the types in the attributes package
type EmailAttributeValue = attributes.EmailAttributeValue
type PhoneAttributeValue = attributes.PhoneAttributeValue
type UsernameAttributeValue = attributes.UsernameAttributeValue
type PasswordAttributeValue = attributes.PasswordAttributeValue
type TOTPAttributeValue = attributes.TOTPAttributeValue
type GitHubAttributeValue = attributes.GitHubAttributeValue
type TelegramAttributeValue = attributes.TelegramAttributeValue
type PasskeyAttributeValue = attributes.PasskeyAttributeValue
type YubicoAttributeValue = attributes.YubicoAttributeValue
type EntitlementSetAttributeValue = attributes.EntitlementSetAttributeValue
type Entitlement = attributes.Entitlement
type EffectType = attributes.EffectType
type OidcAttributeValue = attributes.OidcAttributeValue
type DeviceAttributeValue = attributes.DeviceAttributeValue

// Constants for EffectType
const (
	EffectTypeAllow EffectType = attributes.EffectTypeAllow
	EffectTypeDeny  EffectType = attributes.EffectTypeDeny
)
