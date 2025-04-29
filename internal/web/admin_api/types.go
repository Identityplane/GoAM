package admin_api

// FlowWithYAML represents a flow in the API with its YAML content
type FlowWithYAML struct {
	FlowId string `json:"flow_id"`
	Route  string `json:"route"`
	Realm  string `json:"realm"`
	Tenant string `json:"tenant"`
	YAML   string `json:"yaml"`
}

// FlowPatch represents a partial update to a flow
// Note: FlowId cannot be changed after creation
type FlowPatch struct {
	Route *string `json:"route,omitempty"`
	YAML  *string `json:"yaml,omitempty"`
}

// NodeInfo represents a node definition in the API response
type NodeInfo struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	Type                 string   `json:"type"`
	PossibleResultStates []string `json:"possible_result_states"`
}
