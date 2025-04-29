package admin_api

// FlowPatch represents a partial update to a flow
// Note: FlowId cannot be changed after creation
type FlowPatch struct {
	Route *string `json:"route,omitempty"`
}

// NodeInfo represents a node definition in the API response
type NodeInfo struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	Type                 string   `json:"type"`
	PossibleResultStates []string `json:"possible_result_states"`
}
