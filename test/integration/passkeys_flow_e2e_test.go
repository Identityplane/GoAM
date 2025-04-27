package integration

import (
	"net/http"
	"testing"
)

func TestPasskeysRegistration(t *testing.T) {

	flow := `flow_id: test_passkeys
route: /test_passkeys
definition:
  name: test_passkeys
  description: test passkeys
  start: init
  nodes:
    init:
      use: init
      next:
        start: setVariable
    setVariable:
      use: setVariable
      custom_config:
        key: username
        value: admin
      next:
        done: registerPasskey
    registerPasskey:
      use: registerPasskey
      next:
        success: finish
    finish:
      use: successResult`

	e := SetupIntegrationTest(t, flow)

	e.GET("/acme/customers/auth/test_passkeys").Expect().
		Status(http.StatusOK).
		Body().Contains("passkeysOptions")

}
