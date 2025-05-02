package admin_api

// FlowPatch represents a partial update to a flow
// Note: FlowId cannot be changed after creation
type FlowPatch struct {
	Route  *string `json:"route,omitempty"`
	Active *bool   `json:"active,omitempty"`
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
