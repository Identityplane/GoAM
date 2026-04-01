package flowse2e

import (
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/test/integration"
	"github.com/stretchr/testify/assert"
)

func TestTelegramLoginFlow_JSONAPI(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")
	var sessionID string

	t.Run("Step 1: Get login options", func(t *testing.T) {
		resp := e.GET("/acme/customers/api/v1/telegram-test").
			WithHeader("Accept", "application/json").
			Expect().
			Status(http.StatusOK).
			JSON()

		sessionID = resp.Object().Value("sessionId").String().Raw()
		assert.NotEmpty(t, sessionID)
		resp.Object().HasValue("currentNode", "passwordOrSocialLogin")
	})

	t.Run("Step 2: Choose Telegram", func(t *testing.T) {
		resp := e.POST("/acme/customers/api/v1/telegram-test").
			WithHeader("Content-Type", "application/json").
			WithJSON(flowRequest{
				SessionID:   sessionID,
				CurrentNode: "passwordOrSocialLogin",
				Responses: map[string]string{
					"option": "social1",
				},
			}).
			Expect().
			Status(http.StatusOK).
			JSON()

		resp.Object().HasValue("currentNode", "telegramLogin")
	})

	t.Run("Step 3: Submit tgAuthResult to telegramLogin node", func(t *testing.T) {

		authResult := "eyJhdXRoX2RhdGUiOjQ5MTM2MDM1NzksImZpcnN0X25hbWUiOiJMdWNhIiwiaGFzaCI6IjI1N2NiMjgwOWJhNTFkYTgzZjBmYTdkZTBlYjI2Nzg5YzdkNzFhNGEwYWRlZmE4Njk3ZDg2ZTZjNDBhOTcyZTMiLCJpZCI6Njc0NTczMTEyMCwicGhvdG9fdXJsIjoiaHR0cHM6Ly90Lm1lL2kvdXNlcnBpYy9BQkMiLCJ1c2VybmFtZSI6Ildob0lzTHVjYSJ9"

		resp := e.POST("/acme/customers/api/v1/telegram-test").
			WithHeader("Content-Type", "application/json").
			WithJSON(flowRequest{
				SessionID:   sessionID,
				CurrentNode: "telegramLogin",
				Responses: map[string]string{
					"tgAuthResult": authResult,
				},
			}).
			Expect().
			Status(http.StatusOK).
			JSON()

		resp.Object().HasValue("currentNode", "successResult")
	})

}
