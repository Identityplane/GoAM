package attributes

import (
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// PasskeyAttributeValue is the attribute value for passkeys
// @description Passkey information
type PasskeyAttributeValue struct {
	RPID               string               `json:"rp_id" example:"example.com"`
	CredentialID       string               `json:"credential_id" example:"1234567890"`
	DisplayName        string               `json:"display_name" example:"John Doe"`
	WebAuthnCredential *webauthn.Credential `json:"webauthn_credential" example:"{}"`
	LastUsedAt         *time.Time           `json:"last_used_at" example:"2024-01-01T00:00:00Z"`
}

// GetIndex returns the index of the passkey attribute value
func (p *PasskeyAttributeValue) GetIndex() string {
	return p.CredentialID
}

// IndexIsSensitive returns whether the index should be omitted from JSON API responses
func (p *PasskeyAttributeValue) IndexIsSensitive() bool {
	return true // Passkey credential IDs are sensitive
}
