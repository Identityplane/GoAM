package model

import (
	"encoding/json"

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

// attributeValueConverter is a function type that converts JSON data to an AttributeValue
type attributeValueConverter func([]byte) (AttributeValue, error)

// attributeValueConverters maps attribute type strings to their conversion functions
var attributeValueConverters = map[string]attributeValueConverter{
	AttributeTypeEmail: func(data []byte) (AttributeValue, error) {
		var val EmailAttributeValue
		err := json.Unmarshal(data, &val)
		return &val, err
	},
	AttributeTypePhone: func(data []byte) (AttributeValue, error) {
		var val PhoneAttributeValue
		err := json.Unmarshal(data, &val)
		return &val, err
	},
	AttributeTypeUsername: func(data []byte) (AttributeValue, error) {
		var val UsernameAttributeValue
		err := json.Unmarshal(data, &val)
		return &val, err
	},
	AttributeTypePassword: func(data []byte) (AttributeValue, error) {
		var val PasswordAttributeValue
		err := json.Unmarshal(data, &val)
		return &val, err
	},
	AttributeTypeTOTP: func(data []byte) (AttributeValue, error) {
		var val TOTPAttributeValue
		err := json.Unmarshal(data, &val)
		return &val, err
	},
	AttributeTypeGitHub: func(data []byte) (AttributeValue, error) {
		var val GitHubAttributeValue
		err := json.Unmarshal(data, &val)
		return &val, err
	},
	AttributeTypeTelegram: func(data []byte) (AttributeValue, error) {
		var val TelegramAttributeValue
		err := json.Unmarshal(data, &val)
		return &val, err
	},
	AttributeTypePasskey: func(data []byte) (AttributeValue, error) {
		var val PasskeyAttributeValue
		err := json.Unmarshal(data, &val)
		return &val, err
	},
	AttributeTypeYubico: func(data []byte) (AttributeValue, error) {
		var val YubicoAttributeValue
		err := json.Unmarshal(data, &val)
		return &val, err
	},
	AttributeTypeEntitlements: func(data []byte) (AttributeValue, error) {
		var val EntitlementSetAttributeValue
		err := json.Unmarshal(data, &val)
		return &val, err
	},
	AttributeTypeOidc: func(data []byte) (AttributeValue, error) {
		var val OidcAttributeValue
		err := json.Unmarshal(data, &val)
		return &val, err
	},
	AttributeTypeDevice: func(data []byte) (AttributeValue, error) {
		var val DeviceAttributeValue
		err := json.Unmarshal(data, &val)
		return &val, err
	},
}

// ConvertMapToAttributeValue converts a map[string]interface{} to an AttributeValue
// based on the attribute type string. Returns nil if the type is not registered or conversion fails.
func ConvertMapToAttributeValue(attrType string, mapValue map[string]interface{}) AttributeValue {
	converter, ok := attributeValueConverters[attrType]
	if !ok {
		return nil
	}

	// Convert map to JSON
	jsonData, err := json.Marshal(mapValue)
	if err != nil {
		return nil
	}

	// Convert JSON to AttributeValue
	attrValue, err := converter(jsonData)
	if err != nil {
		return nil
	}

	return attrValue
}
