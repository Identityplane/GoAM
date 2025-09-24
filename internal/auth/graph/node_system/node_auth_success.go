package node_system

import (
	"fmt"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/pkg/model"
)

var SuccessResultNode = &model.NodeDefinition{
	Name:                 model.NODE_SUCCESS_RESULT,
	PrettyName:           "Authentication Success",
	Description:          "Terminal node that indicates successful authentication and returns user information",
	Category:             "Results",
	Type:                 model.NodeTypeResult,
	RequiredContext:      []string{"user_id"}, // expected to be set by now
	OutputContext:        []string{},
	PossibleResultStates: []string{}, // terminal node
	Run:                  RunAuthSuccessNode,
}

func RunAuthSuccessNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// If we have an auth sucess the user needs to be set in the context
	user, err := node_utils.LoadUserFromContext(state, services)
	if err != nil {
		return nil, fmt.Errorf("user is requried for success result: %w", err)
	}

	// If we have no user we return an error
	if user == nil || user.ID == "" {
		return nil, fmt.Errorf("user id is required for success result")
	}

	// Check if the user is locked
	if user.Status == "locked" || user.Status == "disabled" {
		return nil, fmt.Errorf("user is locked or disabled")
	}

	// Set the result
	state.Result = &model.FlowResult{
		UserID:        user.ID,
		Authenticated: true,
	}

	return &model.NodeResult{
		Condition: "",
		Prompts:   nil,
	}, nil

}
