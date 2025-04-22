package realms

import (
	"os"
	"path/filepath"
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

	flow, err := LoadFlowFromYAML(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, flow)

	assert.Equal(t, "test_login_flow", flow.Name)
	assert.Equal(t, "init", flow.Start)
	assert.Contains(t, flow.Nodes, "askPassword")
	assert.Equal(t, "askPassword", flow.Nodes["askPassword"].Name)
	assert.Equal(t, "done", flow.Nodes["askPassword"].Next["submitted"])
	assert.Equal(t, "Login complete.", flow.Nodes["done"].CustomConfig["message"])
}

func TestLoadFlowsFromDir(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, "flow1.yaml")
	file2 := filepath.Join(dir, "flow2.yaml")

	flow1 := `
name: flow_one
route: /one
start: start
nodes:
  start:
    use: init
    next:
      start: step1
  step1:
    use: successResult
`
	flow2 := `
name: flow_two
route: /two
start: entry
nodes:
  entry:
    use: askUsername
    next:
      submitted: result
  result:
    use: successResult
`

	assert.NoError(t, os.WriteFile(file1, []byte(flow1), 0644))
	assert.NoError(t, os.WriteFile(file2, []byte(flow2), 0644))

	flows, err := LoadFlowsFromDir(dir)
	assert.NoError(t, err)
	assert.Len(t, flows, 2)

	names := map[string]bool{}
	for _, f := range flows {
		names[f.Flow.Name] = true
	}

	assert.Contains(t, names, "flow_one")
	assert.Contains(t, names, "flow_two")
}
