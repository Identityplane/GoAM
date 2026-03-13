package authhandler

import (
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/test/integration"
)

func TestMetadata_GetMetadata(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	tenant := "acme"
	realm := "customers"

	// Ensure there are some flows
	flows, err := service.GetServices().FlowService.ListFlows(tenant, realm)
	if err != nil {
		t.Fatalf("Failed to list flows: %v", err)
	}
	if len(flows) == 0 {
		t.Fatalf("No flows found for testing")
	}

	t.Run("Get Metadata", func(t *testing.T) {
		resp := e.GET("/acme/customers/api/v1/").
			WithHeader("Accept", "application/json").
			Expect().
			Status(http.StatusOK).
			JSON()

		// Validate response structure
		metadata := resp.Object()
		metadata.Value("flows").Array().NotEmpty()
		metadata.Value("realm").Object().
			HasValue("name", "Our beautiful Customers").
			ContainsKey("baseUrl").
			ContainsKey("settings")

		// Check first flow
		flow := metadata.Value("flows").Array().First().Object()
		flow.ContainsKey("id")
		flow.ContainsKey("route")
		flow.ContainsKey("description")
	})
}
