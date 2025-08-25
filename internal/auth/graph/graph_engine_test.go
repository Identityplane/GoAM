package graph

import (
	"testing"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRun_SimpleFlow(t *testing.T) {

	// Setup mock repositories with a user named "alice"
	mockUserRepo := repository.NewMockUserRepository()
	mockRepos := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create a test user with username "alice"
	testUser := &model.User{
		ID:             "test-user-123",
		Tenant:         "acme",
		Realm:          "customers",
		Status:         "active",
		UserAttributes: []model.UserAttribute{},
	}

	// Add username attribute
	testUser.AddAttribute(&model.UserAttribute{
		Type:  model.AttributeTypeUsername,
		Index: "alice",
		Value: model.UsernameAttributeValue{
			Username: "alice",
		},
	})

	// Setup mock expectations for user lookup by username
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeUsername, "alice").Return(testUser, nil)

	flow := &model.FlowDefinition{
		Description: "query_flow",
		Start:       "init",
		Nodes: map[string]*model.GraphNode{
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
	graphResult, err := Run(flow, state, nil, mockRepos)
	assert.NoError(t, err)
	assert.Nil(t, graphResult.Result)
	assert.Equal(t, map[string]string{"username": "text"}, graphResult.Prompts)

	// Step 2: Provide input to askUsername
	inputs := map[string]string{"username": "alice"}
	graphResult, err = Run(flow, state, inputs, mockRepos)
	assert.NoError(t, err)
	assert.NotNil(t, graphResult.Result)
	assert.Nil(t, graphResult.Prompts)

	assert.Equal(t, "done", state.Current)
	assert.Equal(t, "alice", state.Context["username"])
	assert.Equal(t, []string{"init:start", "askUsername:prompted:{\"username\":\"text\"}", "askUsername:submitted", "done"}, state.History)

	// Verify that the mock expectations were met
	mockUserRepo.AssertExpectations(t)
}
