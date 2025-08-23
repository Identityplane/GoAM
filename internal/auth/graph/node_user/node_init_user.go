package node_user

import (
	"time"

	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
)

var InitUserNode = &model.NodeDefinition{
	Name:            "initUser",
	PrettyName:      "Init User",
	Description:     "Initializes an empty user object in the context, or overrides the existing user object with an empty one",
	Category:        "User Management",
	Type:            model.NodeTypeLogic,
	RequiredContext: []string{},
	Run:             RunInitUserNode,
}

func RunInitUserNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	state.User = &model.User{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return model.NewNodeResultWithCondition("success")
}
