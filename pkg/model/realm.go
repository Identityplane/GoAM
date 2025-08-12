package model

// Realm represents the static configuration of a realm, typically loaded from YAML files.
// It includes the realm + tenant identifiers and a multiple FlowWithRoute (for now).
type Realm struct {
	Realm         string            `json:"realm" yaml:"realm" db:"realm"`
	RealmName     string            `json:"realm_name" yaml:"realm_name" db:"realm_name"`
	Tenant        string            `json:"tenant" yaml:"tenant" db:"tenant"`
	BaseUrl       string            `json:"base_url" yaml:"base_url" db:"base_url"`
	RealmSettings map[string]string `json:"realm_settings" yaml:"realm_settings" db:"realm_settings"`
}
