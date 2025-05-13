package model

import "time"

// SigningKey represents a signing key in the system
type SigningKey struct {
	Tenant             string     `json:"tenant"`
	Realm              string     `json:"realm"`
	Kid                string     `json:"kid"`
	Active             bool       `json:"active"`
	Algorithm          string     `json:"algorithm"`
	Implementation     string     `json:"implementation"`
	SigningKeyMaterial string     `json:"-"`
	PublicKeyJWK       string     `json:"public_key_jwk"`
	Created            time.Time  `json:"created"`
	Disabled           *time.Time `json:"disabled,omitempty"`
}
