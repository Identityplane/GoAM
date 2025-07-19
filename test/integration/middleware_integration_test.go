package integration

import (
	"os"
	"testing"

	"github.com/Identityplane/GoAM/internal/web"
)

func TestNotFoundRedirectNoUrl(t *testing.T) {
	e := SetupIntegrationTest(t, "")

	e.GET("/non-existent-path").
		Expect().
		Status(404).
		JSON().
		Object().
		HasValue("error", "not found")
}

func TestNotFoundRedirectWithUrl(t *testing.T) {

	os.Setenv("GOIAM_NOT_FOUND_REDIRECT_URL", "https://example.com/")
	defer os.Unsetenv("GOIAM_NOT_FOUND_REDIRECT_URL")
	e := SetupIntegrationTest(t, "")

	// Recreate router with new config
	Router = web.New()

	e.GET("/non-existent-path").
		Expect().
		Status(303).
		Header("Location").IsEqual("https://example.com/")
}

func TestXForwardedForDisabled(t *testing.T) {
	e := SetupIntegrationTest(t, "")

	e.GET("/info").
		WithHeader("X-Forwarded-For", "1.2.3.4").
		Expect().
		Status(200).
		JSON().
		Object().
		HasValue("user_ip", "0.0.0.0") // Should not use X-Forwarded-For when disabled
}

func TestXForwardedForEnabled(t *testing.T) {

	os.Setenv("GOIAM_USE_X_FORWARDED_FOR", "true")
	defer os.Unsetenv("GOIAM_USE_X_FORWARDED_FOR")
	e := SetupIntegrationTest(t, "")

	// Recreate router with new config
	Router = web.New()

	e.GET("/info").
		WithHeader("X-Forwarded-For", "1.2.3.4").
		Expect().
		Status(200).
		JSON().
		Object().
		HasValue("user_ip", "1.2.3.4") // Should use X-Forwarded-For when enabled
}

func TestServerTiming(t *testing.T) {
	os.Setenv("GOIAM_ENABLE_REQUEST_TIMING", "true")
	defer os.Unsetenv("GOIAM_ENABLE_REQUEST_TIMING")
	e := SetupIntegrationTest(t, "")

	// Recreate router with new config
	Router = web.New()

	e.GET("/info").
		Expect().
		Status(200).
		Header("Server-Timing").NotEmpty()
}
