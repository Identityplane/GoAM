package integration

import (
	"net/http"
	"testing"
)

func TestUsernamePasswordFlow_E2E(t *testing.T) {
	e := SetupIntegrationTest(t, "")

	// ---------- REGISTER ----------
	registerCookie := e.GET("/acme/customers/auth/register").Expect().
		Status(http.StatusOK).
		Cookie("session_id").
		Value()

	// Step 1: submit username
	e.POST("/acme/customers/auth/register").
		WithCookie("session_id", registerCookie.Raw()).
		WithForm(map[string]string{
			"step":     "askUsername",
			"username": "testuser",
		}).
		Expect().
		Status(http.StatusOK)

	// Step 2: submit password
	resp := e.POST("/acme/customers/auth/register").
		WithCookie("session_id", registerCookie.Raw()).
		WithForm(map[string]string{
			"step":     "askPassword",
			"password": "testpass123",
		}).
		Expect().
		Status(http.StatusOK)

	resp.Body().Contains("Registration successful")

	// ---------- LOGIN ----------
	loginCookie := e.GET("/acme/customers/auth/login").Expect().
		Status(http.StatusOK).
		Cookie("session_id").
		Value()

	// Submit username
	e.POST("/acme/customers/auth/login").
		WithCookie("session_id", loginCookie.Raw()).
		WithForm(map[string]string{
			"step":     "askUsername",
			"username": "testuser",
		}).
		Expect().
		Status(http.StatusOK)

	// Submit password
	resp2 := e.POST("/acme/customers/auth/login").
		WithCookie("session_id", loginCookie.Raw()).
		WithForm(map[string]string{
			"step":     "askPassword",
			"password": "testpass123",
		}).
		Expect().
		Status(http.StatusOK)

	resp2.Body().Contains("Login successful")
}
