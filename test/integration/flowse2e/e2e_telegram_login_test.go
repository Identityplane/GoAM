package flowse2e

import (
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/test/integration"
	"github.com/stretchr/testify/assert"
)

func TestHTMLFlow_SuccessLeadToSessionCookie(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")
	var sessionCookie string

	t.Run("Step 1: Get login options", func(t *testing.T) {
		sessionCookie = e.GET("/acme/customers/auth/telegram-test").
			Expect().
			Status(http.StatusOK).Cookie("session_id").Value().Raw()
	})

	t.Run("Step 2: Choose Telegram and get redirect", func(t *testing.T) {
		resp := e.POST("/acme/customers/auth/telegram-test/passwordOrSocialLogin").
			WithFormField("option", "social1").
			WithCookie("session_id", sessionCookie).
			Expect().
			Status(http.StatusSeeOther)

		location := resp.Header("Location").Raw()
		assert.Equal(t, "https://oauth.telegram.org/auth?bot_id=1234567890&origin=http%3A%2F%2Flocalhost%3A8080&return_to=http%3A%2F%2Flocalhost%3A8080%2Facme%2Fcustomers%2Fauth%2Ftelegram-test%3Fcallback%3Dtelegram", location)

	})

	t.Run("Step 3: Redirect back to login callback, node should be returned", func(t *testing.T) {
		resp := e.GET("/acme/customers/auth/telegram-test").
			WithQuery("callback", "telegram").
			WithCookie("session_id", sessionCookie).
			Expect().
			Status(http.StatusOK)

		resp.Body().Contains("telegramLogin")
	})

	t.Run("Step 4: The telegram js callback sends the tgAuthResult to the node", func(t *testing.T) {

		authResult := "eyJhdXRoX2RhdGUiOjQ5MTM2MDM1NzksImZpcnN0X25hbWUiOiJMdWNhIiwiaGFzaCI6IjI1N2NiMjgwOWJhNTFkYTgzZjBmYTdkZTBlYjI2Nzg5YzdkNzFhNGEwYWRlZmE4Njk3ZDg2ZTZjNDBhOTcyZTMiLCJpZCI6Njc0NTczMTEyMCwicGhvdG9fdXJsIjoiaHR0cHM6Ly90Lm1lL2kvdXNlcnBpYy9BQkMiLCJ1c2VybmFtZSI6Ildob0lzTHVjYSJ9"

		resp := e.POST("/acme/customers/auth/telegram-test/telegramLogin").
			WithFormField("tgAuthResult", authResult).
			WithCookie("session_id", sessionCookie).
			Expect().
			Status(http.StatusOK)

		resp.Body().Contains("successResult")
	})

}
