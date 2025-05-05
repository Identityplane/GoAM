package model

import "time"

// ClientSession represents a client session in the system
type ClientSession struct {
	// Tenant is the tenant ID
	Tenant string `json:"tenant"`

	// Realm is the realm ID
	Realm string `json:"realm"`

	// ClientSessionID is the unique identifier for the session
	ClientSessionID string `json:"client_session_id"`

	// ClientID is the ID of the client application
	ClientID string `json:"client_id"`

	// GrantType is the OAuth grant type used for this session
	GrantType string `json:"grant_type"`

	// AccessTokenHash is the hashed access token
	AccessTokenHash string `json:"-"`

	// RefreshTokenHash is the hashed refresh token
	RefreshTokenHash string `json:"-"`

	// AuthCodeHash is the hashed authorization code
	AuthCodeHash string `json:"-"`

	// UserID is the ID of the user associated with this session
	UserID string `json:"user_id"`

	// Scope is the OAuth scope for this session
	Scope string `json:"scope"`

	// CodeChallenge is the code challenge for PKCE flow
	CodeChallenge string `json:"code_challenge"`

	// CodeChallengeMethod is the code challenge method for PKCE flow
	CodeChallengeMethod string `json:"code_challenge_method"`

	// LoginSessionStateJson (json) is the resulting state of the login flow
	LoginSessionJson string `json:"login_session_state"`

	// Created is the timestamp when the session was created
	Created time.Time `json:"created"`

	// Expire is the timestamp when the session expires
	Expire time.Time `json:"expire"`
}
