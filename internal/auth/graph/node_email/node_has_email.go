package node_email

import (
	"errors"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/pkg/model"
)

var HasEmailNode = &model.NodeDefinition{
	Name:                 "hasEmail",
	PrettyName:           "Has Email",
	Description:          "Checks if the user has an email and if it is verified",
	Category:             "Email",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"user"},
	OutputContext:        []string{"has_email", "email", "email_verified"},
	PossibleResultStates: []string{"no", "unverified", "verified"},
	Run:                  RunHasEmailNode,
}

func RunHasEmailNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	user, err := node_utils.TryLoadUserFromContext(state, services)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// If we have no user we turn an error
	if user == nil {
		return model.NewNodeResultWithError(errors.New("no user found"))
	}

	// If we have a user we check if they have an email
	emailAttributes, _, err := model.GetAttributes[model.EmailAttributeValue](user, model.AttributeTypeEmail)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// If we have no email attributes we return no
	if len(emailAttributes) == 0 {
		state.Context["has_email"] = "false"
		return model.NewNodeResultWithCondition("no")
	}

	// Check if we find at least one verified email
	for _, emailAttribute := range emailAttributes {
		if emailAttribute.Verified {

			state.Context["email"] = emailAttribute.Email
			state.Context["has_email"] = "true"
			state.Context["email_verified"] = "true"
			return model.NewNodeResultWithCondition("verified")
		}
	}

	// If we don't find any verified email we return unverified
	state.Context["email"] = emailAttributes[0].Email
	state.Context["has_email"] = "true"
	state.Context["email_verified"] = "false"
	return model.NewNodeResultWithCondition("unverified")
}
