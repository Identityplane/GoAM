package node_email

import (
	"context"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/pkg/model"
)

var CheckEmailAvailableNode = &model.NodeDefinition{
	Name:                 "checkEmailAvailable",
	PrettyName:           "Check Email Availability",
	Description:          "Checks if the email is available for use",
	Category:             "Email",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"email"},
	OutputContext:        []string{"email_available"},
	PossibleResultStates: []string{"available", "taken"},
	Run:                  RunCheckEmailAvailableNode,
}

func RunCheckEmailAvailableNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Try to load the user from the context as we need to check if the email is already in use by a different user
	user, err := node_utils.TryLoadUserFromContext(state, services)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// Put the email in the context for the next node
	email := state.Context["email"]

	// Check if there is already a user that has this email but is a different user
	otherUser, err := services.UserRepo.GetByAttributeIndex(context.Background(), model.AttributeTypeEmail, email)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// If we have a user and it is a different user we return an error
	if otherUser != nil && otherUser.ID != user.ID {
		errorMsg := "Email already in use"
		state.Error = &errorMsg
		return model.NewNodeResultWithCondition("email_taken")
	}

	// If we have no user or the user is the same as the one we are checking we return available
	return model.NewNodeResultWithCondition("available")
}
