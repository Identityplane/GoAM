package unit

import (
	"goiam/internal/auth/graph/yaml"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFlowFromYAML(t *testing.T) {
	yamlContent := `
name: test_login_flow
start: init
nodes:
  init:
    use: init
    next:
      start: askUsername

  askUsername:
    use: askUsername
    next:
      submitted: askPassword

  askPassword:
    use: askPassword
    next:
      submitted: done

  done:
    use: successResult
    custom_config:
      message: Login complete.
`

	tmpFile, err := os.CreateTemp("", "test-flow-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	assert.NoError(t, err)
	assert.NoError(t, tmpFile.Close())

	flow, err := yaml.LoadFlowFromYAML(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, flow)

	assert.Equal(t, "test_login_flow", flow.Name)
	assert.Equal(t, "init", flow.Start)
	assert.Contains(t, flow.Nodes, "askPassword")
	assert.Equal(t, "askPassword", flow.Nodes["askPassword"].Name)
	assert.Equal(t, "done", flow.Nodes["askPassword"].Next["submitted"])
	assert.Equal(t, "Login complete.", flow.Nodes["done"].CustomConfig["message"])
}
