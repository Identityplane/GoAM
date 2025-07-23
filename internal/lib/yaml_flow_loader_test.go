package lib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/stretchr/testify/assert"
)

func TestLoadFlowDefinitonFromString(t *testing.T) {
	// Test a simple flow definition
	yamlContent := `
description: Login or register flow
start: init
nodes:
  askPassword:
    name: askPassword
    use: askPassword
    next:
      submitted: validateUsernamePassword
  askUsername:
    name: askUsername
    use: askUsername
    next:
      submitted: askPassword
  authFailure:
    name: failureResult
    use: failureResult
    next: {}
  authSuccess:
    name: successResult
    use: successResult
    next: {}
  init:
    name: init
    use: init
    next:
      start: askUsername
  validateUsernamePassword:
    name: validateUsernamePassword
    use: validateUsernamePassword
    next:
      fail: askPassword
      locked: authFailure
      success: authSuccess`

	flow, err := LoadFlowDefinitonFromString(yamlContent)
	assert.NoError(t, err)
	assert.NotNil(t, flow)
	assert.Equal(t, "Login or register flow", flow.Description)
	assert.Equal(t, "init", flow.Start)
	assert.Len(t, flow.Nodes, 6)
	assert.Equal(t, "init", flow.Nodes["init"].Name)
	assert.Equal(t, "init", flow.Nodes["init"].Use)
	assert.Equal(t, "askUsername", flow.Nodes["init"].Next["start"])
	assert.Equal(t, "validateUsernamePassword", flow.Nodes["askPassword"].Next["submitted"])
	assert.Equal(t, "authSuccess", flow.Nodes["validateUsernamePassword"].Next["success"])
}

func TestLoadFlowDefinitonsFromDir(t *testing.T) {
	// Test loading flow definitions from a directory
	dir := t.TempDir()

	file1 := filepath.Join(dir, "flow1.yaml")
	file2 := filepath.Join(dir, "flow2.yaml")

	flow1 := `
description: Flow one
start: init
nodes:
  init:
    name: init
    use: init
    next:
      start: end
  end:
    name: successResult
    use: successResult
    next: {}`

	flow2 := `
description: Flow two
start: init
nodes:
  init:
    name: init
    use: init
    next:
      start: end
  end:
    name: successResult
    use: successResult
    next: {}`

	assert.NoError(t, os.WriteFile(file1, []byte(flow1), 0644))
	assert.NoError(t, os.WriteFile(file2, []byte(flow2), 0644))

	flows, err := LoadFlowDefinitonsFromDir(dir)
	assert.NoError(t, err)
	assert.Len(t, flows, 2)

	descriptions := map[string]bool{}
	for _, f := range flows {
		descriptions[f.Description] = true
	}

	assert.Contains(t, descriptions, "Flow one")
	assert.Contains(t, descriptions, "Flow two")
}

func TestConvertFlowToYAML(t *testing.T) {
	// Test converting a flow definition back to YAML
	flow := &model.FlowDefinition{
		Description: "test flow",
		Start:       "init",
		Nodes: map[string]*model.GraphNode{
			"init": {
				Name: "init",
				Use:  "init",
				Next: map[string]string{
					"start": "end",
				},
			},
			"end": {
				Name: "end",
				Use:  "successResult",
				Next: map[string]string{},
			},
		},
	}

	yamlStr, err := ConvertFlowToYAML(flow)
	assert.NoError(t, err)
	assert.NotEmpty(t, yamlStr)
	assert.Contains(t, yamlStr, "description: test flow")
	assert.Contains(t, yamlStr, "start: init")
}

func TestLoadFlowWithCustomConfig(t *testing.T) {
	yamlContent := `
description: Test flow with custom config
start: init
nodes:
  init:
    name: init
    use: init
    next:
      start: setVariable
  setVariable:
    name: setVariable
    use: setVariable
    next:
      done: loadUser
    custom_config:
      key: username
      value: admin
  loadUser:
    name: loadUserByUsername
    use: loadUserByUsername
    next:
      loaded: success
      not_found: failure
  success:
    name: successResult
    use: successResult
    next: {}
  failure:
    name: failureResult
    use: failureResult
    next: {}`

	flow, err := LoadFlowDefinitonFromString(yamlContent)
	assert.NoError(t, err)
	assert.NotNil(t, flow)
	assert.Equal(t, "Test flow with custom config", flow.Description)
	assert.Equal(t, "init", flow.Start)
	assert.Len(t, flow.Nodes, 5)

	// Verify custom configuration
	setVariableNode := flow.Nodes["setVariable"]
	assert.NotNil(t, setVariableNode)
	assert.Equal(t, "setVariable", setVariableNode.Name)
	assert.Equal(t, "setVariable", setVariableNode.Use)

	// Verify custom config values
	assert.Equal(t, "username", setVariableNode.CustomConfig["key"])
	assert.Equal(t, "admin", setVariableNode.CustomConfig["value"])

	// Verify next nodes
	assert.Equal(t, "loadUser", setVariableNode.Next["done"])
	assert.Equal(t, "success", flow.Nodes["loadUser"].Next["loaded"])
	assert.Equal(t, "failure", flow.Nodes["loadUser"].Next["not_found"])
}
