package integration_admin_api

import (
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/internal/auth/graph"
	"github.com/Identityplane/GoAM/test/integration"
	"github.com/gavv/httpexpect/v2"
)

// Loads the /admin/nodes endpoint and checks if the nodes are present with the correct fields
func TestNodesAPI_E2E(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")
	nodeDefinitionName := "passwordOrSocialLogin"
	actualNode := graph.GetNodeDefinitionByName(nodeDefinitionName)

	// Act
	response := e.GET("/admin/nodes").
		Expect().
		Status(http.StatusOK).JSON()

	nodes := response.Array()
	nodes.Length().Gt(0)

	// Assert
	// Find node with passwordOrSocialLogin
	passwordOrSocialLoginNode := nodes.Filter(func(index int, value *httpexpect.Value) bool {
		return value.Object().Value("use").String().Raw() == nodeDefinitionName
	})

	passwordOrSocialLoginNode.Length().Gt(0)
	node := passwordOrSocialLoginNode.First()

	// Check the fields of the first matching node
	node.Object().
		HasValue("use", actualNode.Name).
		HasValue("prettyName", actualNode.PrettyName).
		HasValue("type", string(actualNode.Type)).
		HasValue("category", actualNode.Category).
		HasValue("description", actualNode.Description)

	node.Object().Value("requiredContext").Array()
	node.Object().Value("outputContext").Array()
	node.Object().Value("possibleResultStates").Array()
	node.Object().Value("customConfigOptions").Object()

}
