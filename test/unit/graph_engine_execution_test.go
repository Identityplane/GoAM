package unit

import (
	"goiam/internal/auth/graph"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun_SimpleInitToSuccess(t *testing.T) {
	// Build a super simple flow: init -> successResult
	flow := &graph.FlowDefinition{
		Name:  "simple_flow",
		Start: "init",
		Nodes: map[string]*graph.GraphNode{
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

	state := graph.InitFlow(flow)
	assert.Equal(t, "init", state.Current)

	prompts, result, err := graph.Run(flow, state, nil)
	assert.NoError(t, err)
	assert.Nil(t, prompts)
	assert.NotNil(t, result)

	assert.Equal(t, "done", result.Name)
	assert.Equal(t, "done", state.Current)
	assert.Equal(t, []string{"init:start", "done"}, state.History)
}

func TestRun_InitQueryToSuccess(t *testing.T) {
	flow := &graph.FlowDefinition{
		Name:  "query_flow",
		Start: "init",
		Nodes: map[string]*graph.GraphNode{
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

	state := graph.InitFlow(flow)

	// Step 1: Init → askUsername
	prompts, result, err := graph.Run(flow, state, nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, map[string]string{"username": "text"}, prompts)

	// Step 2: Provide input to askUsername
	inputs := map[string]string{"username": "alice"}
	prompts, result, err = graph.Run(flow, state, inputs)
	assert.NoError(t, err)
	assert.Nil(t, prompts)
	assert.NotNil(t, result)

	assert.Equal(t, "done", result.Name)
	assert.Equal(t, "done", state.Current)
	assert.Equal(t, "alice", state.Context["username"])
	assert.Equal(t, []string{"init:start", "askUsername", "askUsername:submitted:{\"username\":\"alice\"}", "done"}, state.History)
}
