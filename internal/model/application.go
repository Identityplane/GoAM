package model

import "time"

// Application represents an OAuth2 client application
type Application struct {
	Tenant          string    `json:"tenant"`
	Realm           string    `json:"realm"`
	ClientId        string    `json:"client_id"`
	ClientSecret    string    `json:"-"`
	Confidential    bool      `json:"confidential"`
	ConsentRequired bool      `json:"consent_required"`
	Description     string    `json:"description"`
	AllowedScopes   []string  `json:"allowed_scopes"`
	AllowedFlows    []string  `json:"allowed_flows"`
	RedirectUris    []string  `json:"redirect_uris"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
