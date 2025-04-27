package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFlowFromYAML(t *testing.T) {
	yamlContent := `
flow_id: test_login_flow
route: /login
active: yes

definition:
  name: test_login_flow
  description: test login flow
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

	assert.Equal(t, "test_login_flow", flow.Id)
	assert.Equal(t, "/login", flow.Route)
	assert.True(t, flow.Active)
	assert.Equal(t, "test_login_flow", flow.Definition.Name)
	assert.Equal(t, "init", flow.Definition.Start)
	assert.Contains(t, flow.Definition.Nodes, "askPassword")
	assert.Equal(t, "askPassword", flow.Definition.Nodes["askPassword"].Name)
	assert.Equal(t, "done", flow.Definition.Nodes["askPassword"].Next["submitted"])
	assert.Equal(t, "Login complete.", flow.Definition.Nodes["done"].CustomConfig["message"])
}

func TestLoadFlowsFromDir(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, "flow1.yaml")
	file2 := filepath.Join(dir, "flow2.yaml")

	flow1 := `
flow_id: flow_one
route: /one
active: yes

definition:
  name: flow_one
  description: flow one
  start: init
  nodes:
    init:
      use: init
`
	flow2 := `
flow_id: flow_two
route: /two
active: yes

definition:
  name: flow_two
  description: flow two
  start: init
  nodes:
    init:
      use: init
`

	assert.NoError(t, os.WriteFile(file1, []byte(flow1), 0644))
	assert.NoError(t, os.WriteFile(file2, []byte(flow2), 0644))

	flows, err := LoadFlowsFromDir(dir)
	assert.NoError(t, err)
	assert.Len(t, flows, 2)

	names := map[string]bool{}
	for _, f := range flows {
		names[f.Definition.Name] = true
	}

	assert.Contains(t, names, "flow_one")
	assert.Contains(t, names, "flow_two")
}
