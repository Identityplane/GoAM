package node_passkeys

import (
	"fmt"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/pkg/model"
)

var PasskeysCheckUserRegistered = &model.NodeDefinition{
	Name:                 "checkPasskeyRegistered",
	PrettyName:           "Check Passkey Registration",
	Description:          "Checks if a user has already registered a passkey for passwordless authentication",
	Category:             "Passkeys",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"username"},
	PossiblePrompts:      nil,
	OutputContext:        []string{},
	PossibleResultStates: []string{"registered", "not_registered", "user_not_found"},
	Run:                  RunCheckUserHasPasskeyNode,
}

func RunCheckUserHasPasskeyNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Try to load user from context, if we cannot load user, return user_not_found
	user, err := node_utils.TryLoadUserFromContext(state, services)
	if err != nil {
		return nil, fmt.Errorf("cannot load user: %w", err)
	}

	// If user is nil, return user_not_found
	if user == nil {
		return model.NewNodeResultWithCondition("user_not_found")
	}

	// Get passkeys from user attributes
	passkeys, _, err := model.GetAttributes[model.PasskeyAttributeValue](user, model.AttributeTypePasskey)
	if err != nil {
		return nil, fmt.Errorf("cannot get passkeys: %w", err)
	}

	// If user has at least one passkey, return registered
	if len(passkeys) > 0 {
		return model.NewNodeResultWithCondition("registered")
	}

	// If user has no passkeys, return not_registered
	return model.NewNodeResultWithCondition("not_registered")
}
