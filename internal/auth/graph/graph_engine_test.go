package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun_SimpleInitToSuccess(t *testing.T) {
	// Build a super simple flow: init -> successResult
	flow := &FlowDefinition{
		Name:  "simple_flow",
		Start: "init",
		Nodes: map[string]*GraphNode{
			"init": {
				Name: "init",
				Use:  "init",
				Next: map[string]string{
					"start": "done",
				},
			},
			"done": {
				Name: "done",
				Use:  "successResult",
			},
		},
	}

	state := InitFlow(flow)
	assert.Equal(t, "init", state.Current)

	graphResult, err := Run(flow, state, nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, graphResult.Prompts)
	assert.NotNil(t, graphResult.Result)

	assert.Equal(t, "done", state.Current)
	assert.Equal(t, []string{"init:start", "done"}, state.History)
}

func TestRun_InitQueryToSuccess(t *testing.T) {
	flow := &FlowDefinition{
		Name:  "query_flow",
		Start: "init",
		Nodes: map[string]*GraphNode{
			"init": {
				Name: "init",
				Use:  "init",
				Next: map[string]string{
					"start": "askUsername",
				},
			},
			"askUsername": {
				Name: "askUsername",
				Use:  "askUsername",
				Next: map[string]string{
					"submitted": "done",
				},
			},
			"done": {
				Name: "done",
				Use:  "successResult",
			},
		},
	}

	state := InitFlow(flow)

	// Step 1: Init → askUsername
	graphResult, err := Run(flow, state, nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, graphResult.Result)
	assert.Equal(t, map[string]string{"username": "text"}, graphResult.Prompts)

	// Step 2: Provide input to askUsername
	inputs := map[string]string{"username": "alice"}
	graphResult, err = Run(flow, state, inputs, nil)
	assert.NoError(t, err)
	assert.NotNil(t, graphResult.Result)
	assert.Nil(t, graphResult.Prompts)

	assert.Equal(t, "done", state.Current)
	assert.Equal(t, "alice", state.Context["username"])
	assert.Equal(t, []string{"init:start", "askUsername:prompted:{\"username\":\"text\"}", "askUsername:submitted", "done"}, state.History)
}
