package node_user

import (
	"strings"
	"testing"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCreateUserNode_NoPasswordKeysStillSucceeds(t *testing.T) {
	testRepo, err := repository.NewTestUserRepository("acme", "customers")
	require.NoError(t, err)
	defer testRepo.Close()

	state := &model.AuthenticationSession{
		Context: map[string]string{
			"user_id": "oidc-user-1",
		},
	}
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			"skipSaveUser": "true",
		},
	}

	result, err := RunCreateUserNode(state, node, nil, &model.Repositories{UserRepo: testRepo})
	assert.NoError(t, err)
	assert.Equal(t, "success", result.Condition)
	assert.NotNil(t, state.User)
}

func TestRunCreateUserNode_FailsOnEmptyUsername(t *testing.T) {
	testRepo, err := repository.NewTestUserRepository("acme", "customers")
	require.NoError(t, err)
	defer testRepo.Close()

	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "   ",
		},
	}
	node := &model.GraphNode{CustomConfig: map[string]string{"skipSaveUser": "true"}}

	result, err := RunCreateUserNode(state, node, nil, &model.Repositories{UserRepo: testRepo})
	assert.NoError(t, err)
	assert.Equal(t, "fail", result.Condition)
	assert.Equal(t, "Username is required", *state.Error)
}

func TestRunCreateUserNode_FailsOnShortPassword(t *testing.T) {
	testRepo, err := repository.NewTestUserRepository("acme", "customers")
	require.NoError(t, err)
	defer testRepo.Close()

	state := &model.AuthenticationSession{
		Context: map[string]string{
			"password": "short12",
		},
	}
	node := &model.GraphNode{CustomConfig: map[string]string{"skipSaveUser": "true"}}

	result, err := RunCreateUserNode(state, node, nil, &model.Repositories{UserRepo: testRepo})
	assert.NoError(t, err)
	assert.Equal(t, "fail", result.Condition)
	assert.Equal(t, "Password must be at least 8 characters", *state.Error)
}

func TestRunCreateUserNode_AcceptsMinLengthPassword(t *testing.T) {
	testRepo, err := repository.NewTestUserRepository("acme", "customers")
	require.NoError(t, err)
	defer testRepo.Close()

	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "alice",
			"password": "short123",
		},
	}
	node := &model.GraphNode{CustomConfig: map[string]string{"skipSaveUser": "true"}}

	result, err := RunCreateUserNode(state, node, nil, &model.Repositories{UserRepo: testRepo})
	assert.NoError(t, err)
	assert.Equal(t, "success", result.Condition)
	assert.Nil(t, state.Error)
}

func TestRunCreateUserNode_FailsOnLongPassword(t *testing.T) {
	testRepo, err := repository.NewTestUserRepository("acme", "customers")
	require.NoError(t, err)
	defer testRepo.Close()

	state := &model.AuthenticationSession{
		Context: map[string]string{
			"password": strings.Repeat("a", 65),
		},
	}
	node := &model.GraphNode{CustomConfig: map[string]string{"skipSaveUser": "true"}}

	result, err := RunCreateUserNode(state, node, nil, &model.Repositories{UserRepo: testRepo})
	assert.NoError(t, err)
	assert.Equal(t, "fail", result.Condition)
	assert.Equal(t, "Password must be at most 64 characters", *state.Error)
}
