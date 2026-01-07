package attributes

import "time"

// PhoneAttributeValue is the attribute value for phones
// @description Phone information
type PhoneAttributeValue struct {
	Phone      string     `json:"phone" example:"+1234567890"`
	Verified   bool       `json:"verified" example:"true"`
	VerifiedAt *time.Time `json:"verified_at" example:"2024-01-01T00:00:00Z"`
}

// GetIndex returns the index of the phone attribute value
func (p *PhoneAttributeValue) GetIndex() string {
	return p.Phone
}

// IndexIsSensitive returns whether the index should be omitted from JSON API responses
func (p *PhoneAttributeValue) IndexIsSensitive() bool {
	return false // Phone numbers are not sensitive
}
