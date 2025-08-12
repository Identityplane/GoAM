package admin_api

// FlowPatch represents a partial update to a flow
// Note: FlowId cannot be changed after creation
type FlowPatch struct {
	Route        *string `json:"route,omitempty"`
	Active       *bool   `json:"active,omitempty"`
	DebugAllowed *bool   `json:"debug_allowed,omitempty"`
}

// RealmPatch represents a partial update to a realm
type RealmPatch struct {
	RealmName     *string            `json:"realm_name,omitempty"`
	BaseUrl       *string            `json:"base_url,omitempty"`
	RealmSettings *map[string]string `json:"realm_settings,omitempty"`
}

// NodeInfo represents a node definition in the API response
type NodeInfo struct {
	Use                  string            `json:"use"`
	PrettyName           string            `json:"prettyName"`
	Type                 string            `json:"type"`
	Category             string            `json:"category"`
	RequiredContext      []string          `json:"requiredContext"`
	OutputContext        []string          `json:"outputContext"`
	PossibleResultStates []string          `json:"possibleResultStates"`
	Description          string            `json:"description"`
	CustomConfigOptions  map[string]string `json:"customConfigOptions"`
}
