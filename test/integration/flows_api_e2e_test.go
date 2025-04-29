package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
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
		"id":     flowId,
		"route":  flowRoute,
		"realm":  realm,
		"tenant": tenant,
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
			HasValue("id", flowId)
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
			HasValue("id", flowId)
	})

	// Test updating a flow
	updatedFlowRoute := "test_flow_updated"
	t.Run("Update Flow", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"id":    flowId,
			"route": updatedFlowRoute,
		}

		e.PATCH("/admin/acme/test_realm/flows/test_flow").
			WithJSON(updatePayload).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("route", updatedFlowRoute).
			HasValue("id", flowId)

		// Verify the update
		e.GET("/admin/acme/test_realm/flows/test_flow").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("tenant", tenant).
			HasValue("realm", realm).
			HasValue("route", updatedFlowRoute).
			HasValue("id", flowId)
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

func TestFlowValidation(t *testing.T) {

	// Test flow validation
	e := SetupIntegrationTest(t, "")

	// Happy case - valid flow definition
	validFlow := `name: Valid Flow
description: A valid flow definition
start: node1
nodes:
  node1:
    use: init
    next:
      success: node2
  node2:
    use: successResult`

	e.POST("/admin/acme/test_realm/flows/validate").
		WithText(validFlow).
		WithHeader("Content-Type", "text/yaml").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		HasValue("errors", []interface{}{}).
		HasValue("valid", true)

	// Unhappy case - invalid flow definition
	invalidFlow := `name: Invalid Flow
description: An invalid flow definition
start: non_existent_node
nodes:
  node1:
    use: auth
    next:
      success: non_existent_node`

	e.POST("/admin/acme/test_realm/flows/validate").
		WithText(invalidFlow).
		WithHeader("Content-Type", "text/yaml").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		HasValue("valid", false).
		HasValue("errors", []interface{}{
			map[string]interface{}{
				"startLineNumber": 1,
				"startColumn":     1,
				"endLineNumber":   1,
				"endColumn":       1,
				"message":         "start node 'non_existent_node' not found in nodes",
				"severity":        8,
			},
		})

	// Unhappy case - invalid YAML syntax
	invalidYAML := `name: Invalid YAML
description: A flow with invalid YAML syntax
start: node1
nodes:
  node1:
    use: auth
    next:
      success: node2
  node2:
    use: success
    custom_config:
      message: "Missing closing quote`

	e.POST("/admin/acme/test_realm/flows/validate").
		WithText(invalidYAML).
		WithHeader("Content-Type", "text/yaml").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		HasValue("valid", false).
		HasValue("errors", []interface{}{
			map[string]interface{}{
				"startLineNumber": 1,
				"startColumn":     1,
				"endLineNumber":   1,
				"endColumn":       1,
				"message":         "Invalid YAML format: yaml: line 12: found unexpected end of stream",
				"severity":        8,
			},
		})

	// Unhappy case - missing required fields
	incompleteFlow := `description: A flow with missing required fields
nodes:
  node1:
    use: auth
    next:
      success: node2
  node2:
    use: success`

	e.POST("/admin/acme/test_realm/flows/validate").
		WithText(incompleteFlow).
		WithHeader("Content-Type", "text/yaml").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		HasValue("valid", false).
		HasValue("errors", []interface{}{
			map[string]interface{}{
				"startLineNumber": 1,
				"startColumn":     1,
				"endLineNumber":   1,
				"endColumn":       1,
				"message":         "start node '' not found in nodes",
				"severity":        8,
			},
		})
}

func TestFlowUpdate(t *testing.T) {
	e := SetupIntegrationTest(t, "")

	originalFlowDefYaml := e.GET("/admin/acme/customers/flows/login_auth/definition").
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	// Add some editor metadata too see it it is perserved
	updatedFlowDefYaml := originalFlowDefYaml + `editor:
  nodes:
    askPassword:
      x: 424
      'y': -102
`

	e.PUT("/admin/acme/customers/flows/login_auth/definition").
		WithText(string(updatedFlowDefYaml)).
		WithHeader("Content-Type", "text/yaml").
		Expect().
		Status(http.StatusOK)

	// get the flow again and check the custom config
	flowDefinitionYaml2 := e.GET("/admin/acme/customers/flows/login_auth/definition").
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	assert.Equal(t, updatedFlowDefYaml, flowDefinitionYaml2)
}
