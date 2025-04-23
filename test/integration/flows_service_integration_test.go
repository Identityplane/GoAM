package integration

import (
	"path/filepath"
	"testing"

	"goiam/internal/config"
	"goiam/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestFlowService_Integration(t *testing.T) {
	// Create a new flow service
	svc := service.NewFlowService()

	config.ConfigPath = "./config"

	// Test initialization with real config files
	err := svc.InitFlows()
	assert.NoError(t, err)

	// Test getting existing flows
	expectedFlows := []string{
		"login_auth",
		"test_passkeys_registration",
		"test_passkeys_verify",
		"user_register",
		"unlock_account",
		"username_password_auth",
	}

	for _, flowName := range expectedFlows {
		flow, exists := svc.GetFlowById("acme", "customers", flowName)
		assert.True(t, exists, "Flow %s should exist", flowName)
		assert.NotNil(t, flow, "Flow %s should not be nil", flowName)
		assert.Equal(t, flowName, flow.Flow.Name, "Flow name should match")
	}

	// Test getting a non-existent flow
	flow, exists := svc.GetFlowById("acme", "customers", "non_existent_flow")
	assert.False(t, exists)
	assert.Nil(t, flow)

	// Test listing all flows
	flows, err := svc.ListFlows("acme", "customers")
	assert.NoError(t, err)
	assert.Len(t, flows, len(expectedFlows), "Should list all flows")

	// Verify all expected flows are present
	flowMap := make(map[string]bool)
	for _, flow := range flows {
		flowMap[flow.Flow.Name] = true
	}
	for _, expectedFlow := range expectedFlows {
		assert.True(t, flowMap[expectedFlow], "Flow %s should be in the list", expectedFlow)
	}
}

func TestFlowService_InvalidConfig(t *testing.T) {
	// Create a new flow service
	svc := service.NewFlowService()

	// Temporarily change the config path to a non-existent directory
	originalConfigPath := config.ConfigPath
	config.ConfigPath = filepath.Join(config.ConfigPath, "non_existent")
	defer func() {
		config.ConfigPath = originalConfigPath
	}()

	// Test initialization with invalid config path
	err := svc.InitFlows()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to init flows from config dir")
}
