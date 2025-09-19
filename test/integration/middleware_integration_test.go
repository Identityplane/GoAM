package integration

import (
	"testing"

	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/internal/web"
	"github.com/spf13/viper"
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

	viper.Set("not_found_redirect_url", "https://example.com/")
	defer viper.Set("not_found_redirect_url", "")
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

	viper.Set("forwarding_proxies", 3)
	defer func() {
		viper.Set("forwarding_proxies", 0)
	}()
	e := SetupIntegrationTest(t, "")

	// Recreate router with new config
	Router = web.New()

	// 3 proxy chain taken from the MDM example
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/X-Forwarded-For
	e.GET("/info").
		WithHeader("X-Forwarded-For", "203.0.113.195,2001:db8:85a3:8d3:1319:8a2e:370:7348,198.51.100.178").
		Expect().
		Status(200).
		JSON().
		Object().
		HasValue("user_ip", "203.0.113.195") // Should use X-Forwarded-For when enabled
}

func TestSingleProxyXForwardedFor(t *testing.T) {

	viper.Set("forwarding_proxies", 1)
	defer func() {
		viper.Set("forwarding_proxies", 0)
		config.ServerSettings.ForwardingProxies = 0
	}()
	e := SetupIntegrationTest(t, "")

	// Recreate router with new config
	Router = web.New()

	// In this example the client "192.168.3.3" has sent us an X-Forwarded-For with addresses already populated
	e.GET("/info").
		WithHeader("X-Forwarded-For", "127.0.0.1,192.168.3.3").
		Expect().
		Status(200).
		JSON().
		Object().
		HasValue("user_ip", "192.168.3.3") // Should use X-Forwarded-For when enabled
}

func TestMultipleXForwardedHeaders(t *testing.T) {

	viper.Set("forwarding_proxies", 3)
	defer func() {
		viper.Set("forwarding_proxies", 0)
		config.ServerSettings.ForwardingProxies = 0
	}()
	e := SetupIntegrationTest(t, "")

	// Recreate router with new config
	Router = web.New()

	// In this example the client "192.168.3.3" has sent us an X-Forwarded-For with addresses already populated and we have 3 proxies
	e.GET("/info").
		WithHeader("X-Forwarded-For", "127.0.0.1,192.168.3.3"). // WithHeader is an append underneath
		WithHeader("X-Forwarded-For", "172.16.100.1,13.12.13.1").
		Expect().
		Status(200).
		JSON().
		Object().
		HasValue("user_ip", "192.168.3.3") // Should use X-Forwarded-For when enabled
}

func TestMaliciousXForwardedFor(t *testing.T) {

	viper.Set("forwarding_proxies", 2)
	defer func() {
		viper.Set("forwarding_proxies", 0)
		config.ServerSettings.ForwardingProxies = 0
	}()
	e := SetupIntegrationTest(t, "")

	// Recreate router with new config
	Router = web.New()

	// In this example the client "192.168.3.3" has sent us an X-Forwarded-For with addresses already populated
	// 172.16.100.2 is the second internal proxy
	e.GET("/info").
		WithHeader("X-Forwarded-For", "127.0.0.1,192.168.3.3,172.16.100.2").
		Expect().
		Status(200).
		JSON().
		Object().
		HasValue("user_ip", "192.168.3.3") // Should use X-Forwarded-For when enabled
}

func TestServerTiming(t *testing.T) {
	viper.Set("enable_request_timing", true)
	defer viper.Set("enable_request_timing", false)
	e := SetupIntegrationTest(t, "")

	// Recreate router with new config
	Router = web.New()

	e.GET("/info").
		Expect().
		Status(200).
		Header("Server-Timing").NotEmpty()
}
