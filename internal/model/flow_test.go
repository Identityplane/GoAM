package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const flowYaml = `login:
  tenant: acme
  realm: customers
  id: login
  route: /login
  active: true
  definition_location: login.yaml
`

func TestFlowDeserialization(t *testing.T) {
	// Create a map to hold the YAML data
	var data map[string]Flow

	// Unmarshal the YAML
	err := yaml.Unmarshal([]byte(flowYaml), &data)
	assert.NoError(t, err, "Failed to unmarshal flow YAML")

	// Get the flow from the map
	flow, exists := data["login"]
	assert.True(t, exists, "Flow 'login' not found in YAML")

	// Verify all fields
	assert.Equal(t, "acme", flow.Tenant, "Tenant mismatch")
	assert.Equal(t, "customers", flow.Realm, "Realm mismatch")
	assert.Equal(t, "login", flow.Id, "Id mismatch")
	assert.Equal(t, "/login", flow.Route, "Route mismatch")
	assert.True(t, flow.Active, "Active should be true")
	assert.Equal(t, "login.yaml", flow.DefinitionLocation, "DefinitionLocation mismatch")
}
