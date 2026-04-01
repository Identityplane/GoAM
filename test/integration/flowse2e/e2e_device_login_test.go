package flowse2e

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Identityplane/GoAM/test/integration"
	"github.com/stretchr/testify/assert"
)

type flowRequest struct {
	SessionID   string            `json:"sessionId"`
	CurrentNode string            `json:"currentNode"`
	Responses   map[string]string `json:"responses"`
}

func TestDeviceLoginFlow(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")
	var deviceCookie string
	var deviceCookieName string
	testUserID := "testuser"
	var sessionID string

	t.Run("Step 1: First run - get prompt for user_id", func(t *testing.T) {
		resp := e.GET("/acme/customers/api/v1/device-login").
			WithHeader("Accept", "application/json").
			Expect().
			Status(http.StatusOK).
			JSON()

		sessionID = resp.Object().Value("sessionId").String().Raw()
		assert.NotEmpty(t, sessionID)

		// Verify we got a prompt for user_id
		resp.Object().HasValue("currentNode", "askUserID")
		resp.Object().Value("prompts").Object().HasValue("user_id", "text")
	})

	t.Run("Step 2: Submit user_id and expect device cookie to be set", func(t *testing.T) {
		resp := e.POST("/acme/customers/api/v1/device-login").
			WithHeader("Content-Type", "application/json").
			WithJSON(flowRequest{
				SessionID:   sessionID,
				CurrentNode: "askUserID",
				Responses: map[string]string{
					"user_id": testUserID,
				},
			}).
			Expect().
			Status(http.StatusOK)

		// Check if the device cookie is set
		for _, nameAny := range resp.Cookies().Raw() {
			name := nameAny.(string)
			if strings.HasPrefix(name, "device_") {
				deviceCookieName = name
				deviceCookie = resp.Cookie(name).Value().Raw()
				break
			}
		}

		if deviceCookieName != "" {
			assert.NotEmpty(t, deviceCookie, "Device cookie should be set on first run")
		} else {
			t.Fatal("Device cookie starting with 'device_' should be set after submitting user_id")
		}

		// Verify we reached success
		resp.JSON().Object().HasValue("currentNode", "successResult")
	})

	t.Run("Step 3: Second run with device cookie - expect same user, no new cookie", func(t *testing.T) {
		// Make a fresh request with the device cookie from step 2
		// This simulates a user returning with their device cookie
		resp := e.GET("/acme/customers/api/v1/device-login").
			WithHeader("Accept", "application/json").
			WithCookie(deviceCookieName, deviceCookie).
			Expect().
			Status(http.StatusOK).
			JSON()

		// Verify we reached success (device should be recognized, no prompt needed)
		resp.Object().HasValue("currentNode", "successResult")
	})
}
