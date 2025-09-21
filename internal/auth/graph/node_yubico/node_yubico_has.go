package node_yubico

import (
	"errors"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/pkg/model"
)

var HasYubicoNode = &model.NodeDefinition{
	Name:                 "hasYubicoOtp",
	PrettyName:           "Has Yubico OTP",
	Description:          "Checks if the user has a Yubico OTP linked",
	Category:             "MFA",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"user"},
	PossibleResultStates: []string{model.ResultStateNo, model.ResultStateYes},
	Run:                  RunHasYubicoNode,
}

func RunHasYubicoNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Try to load user from context
	user, err := node_utils.TryLoadUserFromContext(state, services)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// If we have no user we return an error
	if user == nil {
		return model.NewNodeResultWithError(errors.New("user not found in context"))
	}

	// If we have a user we check if they have a yubikey attribute
	yubikeyAttributes, _, err := model.GetAttributes[model.YubicoAttributeValue](user, model.AttributeTypeYubico)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// If we have no yubikey attributes we return no
	if len(yubikeyAttributes) == 0 {
		return model.NewNodeResultWithCondition(model.ResultStateNo)
	}

	// If we have at least one yubikey attribute we return yes
	return model.NewNodeResultWithCondition(model.ResultStateYes)
}
