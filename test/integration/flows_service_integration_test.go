package integration

import (
	"testing"

	"goiam/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestFlowService_Integration(t *testing.T) {
	// Create a new flow service

	SetupIntegrationTest(t, "")

	// Test getting existing flows
	expectedFlows := []string{
		"login_auth",
		"test_passkeys_registration",
		"test_passkeys_verify",
		"user_register",
		"unlock_account",
		"username_password_auth",
	}

	for _, flowId := range expectedFlows {
		flow, exists := service.GetServices().FlowService.GetFlowById("acme", "customers", flowId)
		assert.True(t, exists, "Flow %s should exist", flowId)
		assert.NotNil(t, flow, "Flow %s should not be nil", flowId)
		assert.Equal(t, flowId, flow.Id, "Flow name should match")
	}

	// Test getting a non-existent flow
	flow, exists := service.GetServices().FlowService.GetFlowById("acme", "customers", "non_existent_flow")
	assert.False(t, exists)
	assert.Nil(t, flow)

	// Test listing all flows
	flows, err := service.GetServices().FlowService.ListFlows("acme", "customers")
	assert.NoError(t, err)
	assert.Len(t, flows, len(expectedFlows), "Should list all flows")

	// Verify all expected flows are present in list flows
	flowMap := make(map[string]bool)
	for _, flow := range flows {
		flowMap[flow.Id] = true
	}
	for _, expectedFlow := range expectedFlows {
		assert.True(t, flowMap[expectedFlow], "Flow %s should be in the list", expectedFlow)
	}
}
