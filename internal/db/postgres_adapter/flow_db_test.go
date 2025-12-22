package postgres_adapter

import (
	"testing"

	"github.com/Identityplane/GoAM/pkg/db"

	"github.com/stretchr/testify/require"
)

func TestPostgresFlowDB(t *testing.T) {
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close()

	flowDB, err := NewPostgresFlowDB(conn)
	require.NoError(t, err)

	db.TemplateTestFlowCRUD(t, flowDB)
}
