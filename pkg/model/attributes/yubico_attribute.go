package attributes

// YubicoAttributeValue is the attribute value for Yubico keys
// @description Yubico information
type YubicoAttributeValue struct {
	// @description The public id for the yubikey
	PublicID string `json:"public_id" example:"vvcijgklnrbf"`

	// @description Whether the credential is locked
	Locked bool `json:"locked" example:"false"`

	// @description The number of failed attempts
	FailedAttempts int `json:"failed_attempts" example:"0"`
}

// GetIndex returns the index of the Yubico attribute value
func (y *YubicoAttributeValue) GetIndex() string {
	return y.PublicID
}

// IndexIsSensitive returns whether the index should be omitted from JSON API responses
func (y *YubicoAttributeValue) IndexIsSensitive() bool {
	return true // Yubico public IDs are sensitive
}
