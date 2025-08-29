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

	Index *string `json:"index" db:"index_value" example:""` // the index can be used to lookup a user by attribute, e.g. by social idp login name. Index should be unique per realm
	Type  string  `json:"type" db:"type" example:"password"` // the type of the attribute, e.g. password, email, phone, etc.
	Value any     `json:"value" db:"value"`                  // the value of the attribute, e.g. password, email, phone, etc.

	CreatedAt time.Time `json:"created_at" db:"created_at" example:"2024-01-01T00:00:00Z"` // The time the attribute was created
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" example:"2024-01-01T00:00:00Z"` // The time the attribute was last updated
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
