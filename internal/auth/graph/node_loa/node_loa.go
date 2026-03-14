package node_loa

import (
	"github.com/Identityplane/GoAM/pkg/model"
)

var SetLOA1Node = &model.NodeDefinition{
	Name:                 "setLoa1",
	PrettyName:           "Set LOA to 1",
	Description:          "Sets the Level of Assurance to 1 in the session context",
	Category:             "LOA",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{},
	OutputContext:        []string{"loa"},
	PossibleResultStates: []string{"success"},
	Run: func(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
		state.Context["loa"] = "1"
		return model.NewNodeResultWithCondition("success")
	},
}

var SetLOA2Node = &model.NodeDefinition{
	Name:                 "setLoa2",
	PrettyName:           "Set LOA to 2",
	Description:          "Sets the Level of Assurance to 2 in the session context",
	Category:             "LOA",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{},
	OutputContext:        []string{"loa"},
	PossibleResultStates: []string{"success"},
	Run: func(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
		state.Context["loa"] = "2"
		return model.NewNodeResultWithCondition("success")
	},
}

var GetLOANode = &model.NodeDefinition{
	Name:                 "getLoa",
	PrettyName:           "Get LOA",
	Description:          "Returns the current Level of Assurance from the session context",
	Category:             "LOA",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{},
	OutputContext:        []string{},
	PossibleResultStates: []string{"none", "0", "1", "2", "other"},
	Run: func(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
		loa, ok := state.Context["loa"]
		if !ok || loa == "" {
			return model.NewNodeResultWithCondition("none")
		}

		switch loa {
		case "0":
			return model.NewNodeResultWithCondition("0")
		case "1":
			return model.NewNodeResultWithCondition("1")
		case "2":
			return model.NewNodeResultWithCondition("2")
		default:
			return model.NewNodeResultWithCondition("other")
		}
	},
}
