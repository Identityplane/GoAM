package node_totp

import (
	"errors"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/pkg/model"
)

var HasTOTPNode = &model.NodeDefinition{
	Name:                 "hasTOTP",
	PrettyName:           "Has TOTP",
	Description:          "Checks if the user has a TOTP linked",
	Category:             "TOTP",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"user"},
	PossibleResultStates: []string{"no", "yes"},
	Run:                  RunHasTOTPNode,
}

func RunHasTOTPNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Try to load user from context
	user, err := node_utils.TryLoadUserFromContext(state, services)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// If we have no user we turn an error
	if user == nil {
		return model.NewNodeResultWithError(errors.New("no user found"))
	}

	// If we have a user we check if they have an email
	totpAttributes, _, err := model.GetAttributes[model.TOTPAttributeValue](user, model.AttributeTypeTOTP)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// If we have no email attributes we return no
	if len(totpAttributes) == 0 {
		return model.NewNodeResultWithCondition("no")
	}

	// If we have at least one TOTP attribute we return yes
	return model.NewNodeResultWithCondition("yes")
}
