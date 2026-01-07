package attributes

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

// GetIndex returns the index of the TOTP attribute value
// TOTP uses the secret key as the index
func (t *TOTPAttributeValue) GetIndex() string {
	return t.SecretKey
}

// IndexIsSensitive returns whether the index should be omitted from JSON API responses
func (t *TOTPAttributeValue) IndexIsSensitive() bool {
	return true // TOTP secret keys are sensitive
}
