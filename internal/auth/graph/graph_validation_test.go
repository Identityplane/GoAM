package graph

import (
	"testing"

	"github.com/gianlucafrei/GoAM/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestEngine_ValidMinimalFlow(t *testing.T) {
	flow := &model.FlowDefinition{
		Description: "simple_init_only",
		Start:       "init",
		Nodes: map[string]*model.GraphNode{
			"init": {
				Name: "init",
				Use:  "init",
				Next: map[string]string{
					"start": "end",
				},
			},
			"end": {
				Name: "end",
				Use:  "successResult",
			},
		},
	}

	err := ValidateFlowDefinition(flow)
	assert.NoError(t, err)
}

func TestEngine_MissingStartNode(t *testing.T) {
	flow := &model.FlowDefinition{
		Description: "no_start",
		Start:       "init",
		Nodes:       map[string]*model.GraphNode{},
	}

	err := ValidateFlowDefinition(flow)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start node")
}

func TestEngine_StartNotInit(t *testing.T) {
	flow := &model.FlowDefinition{
		Description: "bad_start_type",
		Start:       "askUsername",
		Nodes: map[string]*model.GraphNode{
			"askUsername": {
				Name: "askUsername",
				Use:  "askUsername",
				Next: map[string]string{
					"submitted": "end",
				},
			},
			"end": {
				Name: "end",
				Use:  "successResult",
			},
		},
	}

	err := ValidateFlowDefinition(flow)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be of type 'init'")
}

func TestEngine_MissingNextOnLogicNode(t *testing.T) {
	flow := &model.FlowDefinition{
		Description: "missing_next",
		Start:       "init",
		Nodes: map[string]*model.GraphNode{
			"init": {
				Name: "init",
				Use:  "init",
				Next: map[string]string{
					"start": "logicStep",
				},
			},
			"logicStep": {
				Name: "logicStep",
				Use:  "validateUsernamePassword",
				// Next is missing here!
			},
		},
	}

	err := ValidateFlowDefinition(flow)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must define a 'Next' map")
}
