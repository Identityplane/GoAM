package model

// RealmConfig represents the static configuration of a realm, typically loaded from YAML files.
// It includes the realm + tenant identifiers and a multiple FlowWithRoute (for now).
type RealmConfig struct {
	Realm  string          `yaml:"realm"`  // e.g. "customers"
	Tenant string          `yaml:"tenant"` // e.g. "acme"
	Flows  []FlowWithRoute `yaml:"flows"`  // now supports multiple flows
}

type FlowWithRoute struct {
	Route string          // e.g. "/login"
	Flow  *FlowDefinition // pre-loaded flow definition
}
