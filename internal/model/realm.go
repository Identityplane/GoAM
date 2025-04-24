package model

// Realm represents the static configuration of a realm, typically loaded from YAML files.
// It includes the realm + tenant identifiers and a multiple FlowWithRoute (for now).
type Realm struct {
	Realm     string // e.g. "customers"
	RealmName string // e.g. "Our Customers"
	Tenant    string // e.g. "acme"
}
