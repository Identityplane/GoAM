package postgres_adapter

import (
	"context"
	"testing"

	"goiam/internal/db"

	"github.com/stretchr/testify/require"
)

func TestPostgresApplicationDB(t *testing.T) {
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close(context.Background())

	appDB, err := NewPostgresApplicationDB(conn)
	require.NoError(t, err)

	db.TemplateTestApplicationCRUD(t, appDB)
}
