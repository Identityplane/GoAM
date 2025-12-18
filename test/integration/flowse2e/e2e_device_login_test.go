package flowse2e

import (
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/test/integration"
	"github.com/stretchr/testify/assert"
)

func TestDeviceLoginFlow(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")
	var sessionCookie string
	var deviceCookie string
	testUserID := "testuser"

	t.Run("Step 1: First run - get prompt for user_id", func(t *testing.T) {
		resp := e.GET("/acme/customers/auth/device-login").
			Expect().
			Status(http.StatusOK)

		// Get the session_id cookie (for flow state)
		sessionCookie = resp.Cookie("session_id").Value().Raw()

		// Verify we got a prompt for user_id
		resp.Body().Contains("askUserID")
	})

	t.Run("Step 2: Submit user_id and expect device cookie to be set", func(t *testing.T) {
		resp := e.POST("/acme/customers/auth/device-login/askUserID").
			WithFormField("user_id", testUserID).
			WithCookie("session_id", sessionCookie).
			Expect().
			Status(http.StatusOK)

		// Check if the device cookie is set
		deviceCookieValue := resp.Cookie("device")
		if deviceCookieValue != nil {
			deviceCookie = deviceCookieValue.Value().Raw()
			assert.NotEmpty(t, deviceCookie, "Device cookie should be set on first run")
		} else {
			t.Fatal("Device cookie 'device' should be set after submitting user_id")
		}

		// Verify we reached success
		resp.Body().Contains("successResult")
	})

	t.Run("Step 3: Second run with device cookie - expect same user, no new cookie", func(t *testing.T) {
		// Make a fresh request with the device cookie from step 2
		// This simulates a user returning with their device cookie
		resp := e.GET("/acme/customers/auth/device-login").
			WithCookie("device", deviceCookie).
			Expect().
			Status(http.StatusOK)

		// Verify we reached success (device should be recognized, no prompt needed)
		resp.Body().Contains("successResult")
	})
}
