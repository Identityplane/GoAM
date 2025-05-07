package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORS_Integration(t *testing.T) {
	e := SetupIntegrationTest(t, "")

	// Test /authorize endpoint - no CORS
	t.Run("OPTIONS on authorize endpoint no CORS", func(t *testing.T) {

		// Test OPTIONS request
		resp := e.OPTIONS("/internal/internal/oauth2/authorize?").
			Expect().
			Status(http.StatusNotFound)

		assert.Empty(t, resp.Header("Access-Control-Allow-Origin").Raw(), "CORS header should not exist")
	})

	// Test /admin/nodes endpoint - CORS enabled with OPTIONS
	t.Run("admin nodes endpoint with CORS and OPTIONS", func(t *testing.T) {
		// Test regular request
		resp := e.GET("/admin/realms").
			WithHeader("Origin", "https://example.com").
			Expect().
			Status(200)

		assert.NotEmpty(t, resp.Header("Access-Control-Allow-Origin").Raw(), "CORS header should exist")

		// Test OPTIONS request
		resp = e.OPTIONS("/admin/realms").
			WithHeader("Origin", "https://example.com").
			WithHeader("Access-Control-Request-Method", "GET").
			Expect().
			Status(200)

		assert.NotEmpty(t, resp.Header("Access-Control-Allow-Origin").Raw(), "CORS header should exist")
		assert.NotEmpty(t, resp.Header("Access-Control-Allow-Methods").Raw(), "CORS methods header should exist")
		assert.NotEmpty(t, resp.Header("Access-Control-Allow-Headers").Raw(), "CORS headers header should exist")
	})
}
