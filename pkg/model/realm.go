package model

// Realm represents the static configuration of a realm, typically loaded from YAML files.
// It includes the realm + tenant identifiers and a multiple FlowWithRoute (for now).
type Realm struct {
	Realm         string            `json:"realm"`          // e.g. "customers"
	RealmName     string            `json:"realm_name"`     // e.g. "Our Customers"
	Tenant        string            `json:"tenant"`         // e.g. "acme"
	BaseUrl       string            `json:"base_url"`       // e.g. "https://acme.com"
	RealmSettings map[string]string `json:"realm_settings"` // e.g. {"theme": "dark"}
}
