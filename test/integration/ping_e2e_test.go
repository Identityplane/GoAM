package integration

import (
	"net/http"
	"testing"
)

func TestPingE2E(t *testing.T) {

	e := SetupIntegrationTest(t)

	e.GET("/ping").
		Expect().
		Status(http.StatusOK).
		Body().Equal("pong")
}
