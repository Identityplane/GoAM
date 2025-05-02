package sqlite_adapter

import (
	"context"
	"goiam/internal/db"
	"goiam/internal/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlowCRUD(t *testing.T) {
	sqldb := setupTestDB(t)
	flowDB, err := NewFlowDB(sqldb)
	require.NoError(t, err)

	db.TemplateTestFlowCRUD(t, flowDB)
}

func TestFlowUniqueConstraints(t *testing.T) {
	sqldb := setupTestDB(t)
	flowDB, err := NewFlowDB(sqldb)
	require.NoError(t, err)

	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"

	// Create first flow
	flow1 := model.Flow{
		Tenant:        testTenant,
		Realm:         testRealm,
		Id:            "flow1",
		Route:         "/test1",
		Active:        true,
		DefintionYaml: "test1: yaml",
	}
	err = flowDB.CreateFlow(ctx, flow1)
	require.NoError(t, err)

	// Try to create another flow with same route (should fail)
	flow2 := model.Flow{
		Tenant:        testTenant,
		Realm:         testRealm,
		Id:            "flow2",
		Route:         "/test1", // same route
		Active:        true,
		DefintionYaml: "test2: yaml",
	}
	err = flowDB.CreateFlow(ctx, flow2)
	require.Error(t, err)

	// Try to create another flow with same id (should fail)
	flow3 := model.Flow{
		Tenant:        testTenant,
		Realm:         testRealm,
		Id:            "flow1", // same id
		Route:         "/test3",
		Active:        true,
		DefintionYaml: "test3: yaml",
	}
	err = flowDB.CreateFlow(ctx, flow3)
	require.Error(t, err)
}
