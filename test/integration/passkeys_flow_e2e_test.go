package integration

import (
	"context"
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"
)

func TestPasskeysRegistration(t *testing.T) {

	flow := `
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
      done: loadUser
  loadUser:
    use: loadUserByUsername
    next:
      loaded: registerPasskey
  registerPasskey:
    use: registerPasskey
    next:
      success: finish
  finish:
    use: successResult`

	e := SetupIntegrationTest(t, flow)

	service.GetServices().UserService.CreateUser(context.Background(), "acme", "customers", model.User{
		Username: "admin",
	})

	e.GET("/acme/customers/auth/test_flow").Expect().
		Status(http.StatusOK).
		Body().Contains("passkeysOptions")

}
