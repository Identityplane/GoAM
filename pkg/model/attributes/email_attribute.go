package attributes

import "time"

// EmailAttributeValue is the attribute value for emails
// @description Email information
type EmailAttributeValue struct {
	Email      string     `json:"email" example:"john.doe@example.com"`
	Verified   bool       `json:"verified" example:"true"`
	VerifiedAt *time.Time `json:"verified_at" example:"2024-01-01T00:00:00Z"`

	OtpFailedAttempts int  `json:"otp_failed_attempts" example:"0"`
	OtpLocked         bool `json:"otp_locked" example:"false"`
}

// GetIndex returns the index of the email attribute value
func (e *EmailAttributeValue) GetIndex() string {
	return e.Email
}

// IndexIsSensitive returns whether the index should be omitted from JSON API responses
func (e *EmailAttributeValue) IndexIsSensitive() bool {
	return false // Email addresses are not sensitive
}
