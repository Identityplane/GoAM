package integration_admin_api

import (
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/test/integration"
	"github.com/gavv/httpexpect/v2"
)

func TestNodesAPI_E2E(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	// Act
	response := e.GET("/admin/nodes").
		Expect().
		Status(http.StatusOK).JSON()

	nodes := response.Array()
	nodes.Length().Gt(0)

	// Assert
	// Find node with passwordOrSocialLogin
	passwordOrSocialLoginNode := nodes.Filter(func(index int, value *httpexpect.Value) bool {
		return value.Object().Value("use").String().Raw() == "passwordOrSocialLogin"
	})

	passwordOrSocialLoginNode.Length().Gt(0)
	node := passwordOrSocialLoginNode.First()

	// Check the fields of the first matching node
	node.Object().
		HasValue("use", "passwordOrSocialLogin").
		HasValue("prettyName", "Password or Social Login").
		HasValue("type", "queryWithLogic").
		HasValue("category", "").
		HasValue("description", "This node is used to login with password or social login")

	node.Object().
		Value("requiredContext").Array().ConsistsOf("")

	node.Object().
		Value("outputContext").Array().ConsistsOf("username", "password")

	node.Object().
		Value("possibleResultStates").Array().ConsistsOf(
		"password",
		"forgotPassword",
		"passkey",
		"social1",
		"social2",
		"social3",
	)

	node.Object().
		Value("customConfigOptions").Object().
		ContainsKey("useEmail").
		ContainsKey("showForgotPassword").
		ContainsKey("showPasskeys").
		ContainsKey("social1").
		ContainsKey("social2").
		ContainsKey("social1Provider").
		ContainsKey("social2Provider")

}
