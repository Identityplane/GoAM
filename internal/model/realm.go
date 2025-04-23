package model

// RealmConfig represents the static configuration of a realm, typically loaded from YAML files.
// It includes the realm + tenant identifiers and a multiple FlowWithRoute (for now).
type RealmConfig struct {
	Realm  string // e.g. "customers"
	Tenant string // e.g. "acme"
	//Flows  []FlowWithRoute `yaml:"flows"`  // now supports multiple flows
	ActiveFlows []string // list of flow names
}
