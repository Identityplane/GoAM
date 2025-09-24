package integration

import (
	"testing"

	"github.com/Identityplane/GoAM/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestFlowService_Integration(t *testing.T) {
	// Create a new flow service
	SetupIntegrationTest(t, "")

	// Test getting a non-existent flow
	flow, exists := service.GetServices().FlowService.GetFlowById("acme", "customers", "non_existent_flow")
	assert.False(t, exists)
	assert.Nil(t, flow)

	// Test listing all flows - just check that there are flows available
	flows, err := service.GetServices().FlowService.ListFlows("acme", "customers")
	assert.NoError(t, err)
	assert.Greater(t, len(flows), 1, "Should have more than 1 flow available")

	// Verify that flows have valid IDs
	for _, flow := range flows {
		assert.NotEmpty(t, flow.Id, "Flow should have a valid ID")
		assert.NotNil(t, flow, "Flow should not be nil")
	}
}
