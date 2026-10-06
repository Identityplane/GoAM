package node_username

import (
	"testing"

	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestRunAskUsernameNode_PromptWhenMissing(t *testing.T) {
	state := &model.AuthenticationSession{Context: map[string]string{}}
	node := &model.GraphNode{Name: "askUsername", Use: "askUsername"}

	result, err := RunAskUsernameNode(state, node, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "text", result.Prompts["username"])
	assert.Empty(t, result.Condition)
	assert.Nil(t, state.Error)
}

func TestRunAskUsernameNode_RejectsEmptyAndWhitespace(t *testing.T) {
	state := &model.AuthenticationSession{Context: map[string]string{}}
	node := &model.GraphNode{Name: "askUsername", Use: "askUsername"}

	for _, username := range []string{"", "   "} {
		result, err := RunAskUsernameNode(state, node, map[string]string{"username": username}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "text", result.Prompts["username"])
		assert.Empty(t, result.Condition)
		assert.NotNil(t, state.Error)
		assert.Equal(t, "Username is required", *state.Error)
	}
}

func TestRunAskUsernameNode_TrimsAndSubmits(t *testing.T) {
	state := &model.AuthenticationSession{Context: map[string]string{}}
	node := &model.GraphNode{Name: "askUsername", Use: "askUsername"}

	result, err := RunAskUsernameNode(state, node, map[string]string{"username": "  alice  "}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "submitted", result.Condition)
	assert.Empty(t, result.Prompts)
	assert.Nil(t, state.Error)
	assert.Equal(t, "alice", state.Context["username"])
}
