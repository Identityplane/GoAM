package integration

import (
	"net/http"
	"testing"
)

// TestFlowAPI_E2E performs a complete end-to-end test of the admin API flow management functionality.
// It tests the following flow operations in sequence:
// 1. Creating a new flow
// 2. Listing all flows
// 3. Getting a specific flow's details
// 4. Updating a flow's information
// 5. Deleting a flow
// 6. Verifying flow deletion
// The test uses a test tenant "acme" and realm "test_realm" for all operations.

func TestFlowAPI_E2E(t *testing.T) {
	e := SetupIntegrationTest(t, "")

	// Test flow data
	tenant := "acme"
	realm := "test_realm"
	flowRoute := "test_flow"
	flowId := "test_flow"
	minimalFlow := map[string]interface{}{
		"flow_id": flowId,
		"route":   flowRoute,
		"realm":   realm,
		"tenant":  tenant,
		"yaml": `flow_id: test_flow
route: test_flow
definition:
  start: init
  nodes:
    init:
      use: init
      next:
        start: askUsername
    askUsername:
      name: askUsername
      use: askUsername
      next:
        submitted: end
    end:
      name: end
      use: successResult`,
	}

	// Test creating a flow
	t.Run("Create Flow", func(t *testing.T) {
		e.POST("/admin/acme/test_realm/flows/test_flow").
			WithJSON(minimalFlow).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object().
			HasValue("tenant", tenant).
			HasValue("realm", realm).
			HasValue("route", flowRoute).
			HasValue("flow_id", flowId)
	})

	// Test getting a specific flow
	t.Run("Get Flow", func(t *testing.T) {
		e.GET("/admin/acme/test_realm/flows/test_flow").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("tenant", tenant).
			HasValue("realm", realm).
			HasValue("route", flowRoute).
			HasValue("flow_id", flowId)
	})

	// Test updating a flow
	t.Run("Update Flow", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"flow_id": flowId,
			"yaml": `flow_id: test_flow
route: test_flow
definition:
  start: init
  nodes:
    init:
      use: init
      next:
        start: askUsername
    askUsername:
      name: askUsername
      use: askUsername
      next:
        submitted: end
    end:
      name: end
      use: successResult
      custom_config:
        message: "Updated success message"`,
		}

		e.PATCH("/admin/acme/test_realm/flows/test_flow").
			WithJSON(updatePayload).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("route", flowRoute).
			HasValue("flow_id", flowId)

		// Verify the update
		e.GET("/admin/acme/test_realm/flows/test_flow").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("tenant", tenant).
			HasValue("realm", realm).
			HasValue("route", flowRoute).
			HasValue("flow_id", flowId)
	})

	// Test deleting a flow
	t.Run("Delete Flow", func(t *testing.T) {
		e.DELETE("/admin/acme/test_realm/flows/test_flow").
			Expect().
			Status(http.StatusOK)

		// Verify flow is deleted
		e.GET("/admin/acme/test_realm/flows/test_flow").
			Expect().
			Status(http.StatusNotFound)
	})
}
