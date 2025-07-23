package db

import (
	"context"
	"testing"

	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TemplateTestFlowCRUD is a parameterized test for basic CRUD operations on flows
func TemplateTestFlowCRUD(t *testing.T, db FlowDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"

	// Create test flow
	testFlow := model.Flow{
		Tenant:         testTenant,
		Realm:          testRealm,
		Id:             "test-flow",
		Route:          "/test",
		Active:         true,
		DefinitionYaml: "test: yaml",
	}

	t.Run("CreateFlow", func(t *testing.T) {
		err := db.CreateFlow(ctx, testFlow)
		assert.NoError(t, err)
	})

	t.Run("GetFlow", func(t *testing.T) {
		flow, err := db.GetFlow(ctx, testTenant, testRealm, testFlow.Id)
		assert.NoError(t, err)
		assert.NotNil(t, flow)
		assert.Equal(t, testFlow.Id, flow.Id)
		assert.Equal(t, testFlow.Route, flow.Route)
		assert.Equal(t, testFlow.DefinitionYaml, flow.DefinitionYaml)
	})

	t.Run("GetFlowByRoute", func(t *testing.T) {
		flow, err := db.GetFlowByRoute(ctx, testTenant, testRealm, testFlow.Route)
		assert.NoError(t, err)
		assert.NotNil(t, flow)
		assert.Equal(t, testFlow.Id, flow.Id)
		assert.Equal(t, testFlow.Route, flow.Route)
	})

	t.Run("UpdateFlow", func(t *testing.T) {
		flow, err := db.GetFlow(ctx, testTenant, testRealm, testFlow.Id)
		require.NoError(t, err)
		require.NotNil(t, flow)

		flow.Route = "/updated"
		flow.DefinitionYaml = "updated: yaml"
		err = db.UpdateFlow(ctx, flow)
		assert.NoError(t, err)

		updatedFlow, err := db.GetFlow(ctx, testTenant, testRealm, testFlow.Id)
		assert.NoError(t, err)
		assert.Equal(t, "/updated", updatedFlow.Route)
		assert.Equal(t, "updated: yaml", updatedFlow.DefinitionYaml)
	})

	t.Run("ListFlows", func(t *testing.T) {
		flows, err := db.ListFlows(ctx, testTenant, testRealm)
		assert.NoError(t, err)
		assert.Len(t, flows, 1)
		assert.Equal(t, testFlow.Id, flows[0].Id)
	})

	t.Run("DeleteFlow", func(t *testing.T) {
		err := db.DeleteFlow(ctx, testTenant, testRealm, testFlow.Id)
		assert.NoError(t, err)

		flow, err := db.GetFlow(ctx, testTenant, testRealm, testFlow.Id)
		assert.NoError(t, err)
		assert.Nil(t, flow)
	})
}
