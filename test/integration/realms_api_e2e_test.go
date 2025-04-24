package integration

import (
	"fmt"
	"net/http"
	"testing"
)

// TestRealmsAPI_E2E performs an end-to-end test of the realms API functionality.
// It tests the following operations:
// 1. Listing all available realms
// 2. Verifying realm data structure
// The test uses the configured realms from the test environment.

func TestRealmsAPI_E2E(t *testing.T) {
	e := SetupIntegrationTest(t, "")

	const expectedRealm = "customers"
	const expectedTenant = "acme"

	t.Run("List Realms", func(t *testing.T) {
		resp := e.GET("/admin/realms").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		// Verify we have at least one realm
		resp.Length().Gt(0)

		// Get the first tenant from the response
		firstTenant := resp.Element(0).Object()
		firstTenant.Value("tenant").String().Equal(expectedTenant)

		// Get the first realm from the tenant
		firstRealm := firstTenant.Value("realms").Array().Element(0).Object()
		firstRealm.Value("realm").String().Equal(expectedRealm)
	})

	t.Run("Get Dashboard", func(t *testing.T) {
		resp := e.GET(fmt.Sprintf("/admin/%s/%s/dashboard", expectedTenant, expectedRealm)).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		// Verify the dashboard response
		resp.Value("user_stats").Object().
			ContainsKey("total_users").
			ContainsKey("active_users")

		// Verify the flows
		resp.Value("flows").Object().
			ContainsKey("total_flows").
			ContainsKey("active_flows")
	})
}
