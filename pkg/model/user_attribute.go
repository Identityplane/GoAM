package model

import (
	"encoding/json"
	"time"
)

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

	UserID string `json:"-" db:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"` // The id of the user who this attribute belongs to
	Tenant string `json:"-" db:"tenant" example:"acme"`                                  // The tenant of the user who this attribute belongs to
	Realm  string `json:"-" db:"realm" example:"customers"`                              // The realm of the user who this attribute belongs to

	Index *string `json:"index,omitempty" db:"index_value" example:""` // the index can be used to lookup a user by attribute, e.g. by social idp login name. Index should be unique per realm
	Type  string  `json:"type" db:"type" example:"password"`           // the type of the attribute, e.g. password, email, phone, etc.
	Value any     `json:"value" db:"value"`                            // the value of the attribute, e.g. password, email, phone, etc.

	CreatedAt time.Time `json:"created_at" db:"created_at" example:"2024-01-01T00:00:00Z"` // The time the attribute was created
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" example:"2024-01-01T00:00:00Z"` // The time the attribute was last updated
}

// MarshalJSON implements custom JSON marshaling to omit the index if it's sensitive
func (ua *UserAttribute) MarshalJSON() ([]byte, error) {
	// Create a type alias to avoid infinite recursion
	type Alias UserAttribute

	// Check if the index is sensitive
	indexIsSensitive := ua.isIndexSensitive()

	// Create a struct for JSON marshaling
	aux := &struct {
		*Alias
		Index *string `json:"index,omitempty"`
	}{
		Alias: (*Alias)(ua),
	}

	// Only include index if it's not sensitive
	if indexIsSensitive {
		aux.Index = nil
	} else {
		aux.Index = ua.Index
	}

	return json.Marshal(aux)
}

// isIndexSensitive checks if the attribute's index should be omitted from JSON responses
func (ua *UserAttribute) isIndexSensitive() bool {
	if ua.Value == nil {
		return false
	}

	// Try to convert the value to AttributeValue interface
	if attrValue, ok := ua.Value.(AttributeValue); ok {
		return attrValue.IndexIsSensitive()
	}

	// If direct conversion fails, try to convert from map[string]interface{} (for database stored values)
	if mapValue, ok := ua.Value.(map[string]interface{}); ok {
		// Convert map to JSON and then try to unmarshal to known attribute types
		jsonData, err := json.Marshal(mapValue)
		if err != nil {
			return false
		}

		// Try each known attribute type based on the Type field
		switch ua.Type {
		case AttributeTypeEmail:
			var val EmailAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				return val.IndexIsSensitive()
			}
		case AttributeTypePhone:
			var val PhoneAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				return val.IndexIsSensitive()
			}
		case AttributeTypeUsername:
			var val UsernameAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				return val.IndexIsSensitive()
			}
		case AttributeTypePassword:
			var val PasswordAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				return val.IndexIsSensitive()
			}
		case AttributeTypeTOTP:
			var val TOTPAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				return val.IndexIsSensitive()
			}
		case AttributeTypeGitHub:
			var val GitHubAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				return val.IndexIsSensitive()
			}
		case AttributeTypeTelegram:
			var val TelegramAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				return val.IndexIsSensitive()
			}
		case AttributeTypePasskey:
			var val PasskeyAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				return val.IndexIsSensitive()
			}
		case AttributeTypeYubico:
			var val YubicoAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				return val.IndexIsSensitive()
			}
		case AttributeTypeEntitlements:
			var val EntitlementSetAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				return val.IndexIsSensitive()
			}
		case AttributeTypeOidc:
			var val OidcAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				return val.IndexIsSensitive()
			}
		case AttributeTypeDevice:
			var val DeviceAttributeValue
			if err := json.Unmarshal(jsonData, &val); err == nil {
				return val.IndexIsSensitive()
			}
		}
	}

	return false
}

// Equals compares this attribute with another attribute, checking all fields except timestamps
// This is useful for determining if an attribute has actually changed and needs updating
func (ua *UserAttribute) Equals(other *UserAttribute) bool {
	// Handle nil cases
	if ua == nil && other == nil {
		return true
	}
	if ua == nil || other == nil {
		return false
	}

	// Compare all fields except timestamps
	if ua.ID != other.ID ||
		ua.UserID != other.UserID ||
		ua.Tenant != other.Tenant ||
		ua.Realm != other.Realm ||
		ua.Type != other.Type {
		return false
	}

	// Compare index values (handle nil cases)
	if ua.Index == nil && other.Index == nil {
		// Both are nil, continue to value comparison
	} else if ua.Index == nil || other.Index == nil {
		// One is nil, the other isn't
		return false
	} else if *ua.Index != *other.Index {
		// Both are not nil, compare values
		return false
	}

	// Compare values by converting them to JSON
	if ua.Value == nil && other.Value == nil {
		return true
	}
	if ua.Value == nil || other.Value == nil {
		return false
	}

	// Convert both values to JSON for comparison
	json1, err := json.Marshal(ua.Value)
	if err != nil {
		return false
	}
	json2, err := json.Marshal(other.Value)
	if err != nil {
		return false
	}

	return string(json1) == string(json2)
}
