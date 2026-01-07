package attributes

import "time"

// PasswordAttributeValue is the attribute value for passwords
// @description Password information
type PasswordAttributeValue struct {
	PasswordHash         string     `json:"password_hash" example:"password"`
	Locked               bool       `json:"locked" example:"false"`
	FailedAttempts       int        `json:"failed_attempts" example:"0"`
	LastCorrectTimestamp *time.Time `json:"last_correct_timestamp,omitempty" example:"2024-01-01T00:00:00Z"`
}

// GetIndex returns the index of the password attribute value
// Passwords don't have an index for lookup, so return empty string
func (p *PasswordAttributeValue) GetIndex() string {
	return ""
}

// IndexIsSensitive returns whether the index should be omitted from JSON API responses
func (p *PasswordAttributeValue) IndexIsSensitive() bool {
	return true // Passwords are sensitive (even though they don't have an index)
}
