package node_password

import (
	"strings"
	"testing"

	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestRunAskPasswordNode_PromptWhenMissing(t *testing.T) {
	state := &model.AuthenticationSession{Context: map[string]string{}}
	node := &model.GraphNode{Name: "askPassword", Use: "askPassword"}

	result, err := RunAskPasswordNode(state, node, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "password", result.Prompts["password"])
	assert.Empty(t, result.Condition)
}

func TestRunAskPasswordNode_RejectsEmpty(t *testing.T) {
	state := &model.AuthenticationSession{Context: map[string]string{}}
	node := &model.GraphNode{Name: "askPassword", Use: "askPassword"}

	result, err := RunAskPasswordNode(state, node, map[string]string{"password": ""}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "password", result.Prompts["password"])
	assert.NotNil(t, state.Error)
	assert.Equal(t, "Password is required", *state.Error)
}

func TestRunAskPasswordNode_MinLengthOnlyWhenConfigured(t *testing.T) {
	state := &model.AuthenticationSession{Context: map[string]string{}}

	// Login-style: no minLength, short password is accepted
	loginNode := &model.GraphNode{CustomConfig: map[string]string{}}
	result, err := RunAskPasswordNode(state, loginNode, map[string]string{"password": "short"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "submitted", result.Condition)
	assert.Equal(t, "short", state.Context["password"])

	// Registration-style: minLength 8
	state = &model.AuthenticationSession{Context: map[string]string{}}
	registerNode := &model.GraphNode{CustomConfig: map[string]string{"minLength": "8"}}
	result, err = RunAskPasswordNode(state, registerNode, map[string]string{"password": "short12"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "password", result.Prompts["password"])
	assert.Equal(t, "Password must be at least 8 characters", *state.Error)

	result, err = RunAskPasswordNode(state, registerNode, map[string]string{"password": "short123"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "submitted", result.Condition)
	assert.Nil(t, state.Error)
	assert.Equal(t, "short123", state.Context["password"])
}

func TestRunAskPasswordNode_MaxLengthWhenConfigured(t *testing.T) {
	state := &model.AuthenticationSession{Context: map[string]string{}}
	registerNode := &model.GraphNode{CustomConfig: map[string]string{
		"minLength": "8",
		"maxLength": "64",
	}}

	tooLong := strings.Repeat("a", 65)
	result, err := RunAskPasswordNode(state, registerNode, map[string]string{"password": tooLong}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "password", result.Prompts["password"])
	assert.Equal(t, "Password must be at most 64 characters", *state.Error)

	ok := strings.Repeat("a", 64)
	result, err = RunAskPasswordNode(state, registerNode, map[string]string{"password": ok}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "submitted", result.Condition)
	assert.Nil(t, state.Error)
}

func TestRunAskUsernamePasswordNode_RejectsEmpty(t *testing.T) {
	state := &model.AuthenticationSession{Context: map[string]string{}}
	node := &model.GraphNode{Name: "askUsernamePassword", Use: "askUsernamePassword"}

	result, err := RunAskUsernamePasswordNode(state, node, map[string]string{
		"username": "",
		"password": "secret",
	}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Username is required", *state.Error)
	assert.NotEmpty(t, result.Prompts)

	result, err = RunAskUsernamePasswordNode(state, node, map[string]string{
		"username": "alice",
		"password": "",
	}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Password is required", *state.Error)

	result, err = RunAskUsernamePasswordNode(state, node, map[string]string{
		"username": "  alice  ",
		"password": "short", // login: no min length
	}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "submitted", result.Condition)
	assert.Equal(t, "alice", state.Context["username"])
	assert.Equal(t, "short", state.Context["password"])
}
