package postgres_adapter

import (
	"context"
	"testing"

	"goiam/internal/db"

	"github.com/stretchr/testify/require"
)

func TestPostgresClientSessionDB(t *testing.T) {
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close(context.Background())

	sessionDB, err := NewPostgresClientSessionDB(conn)
	require.NoError(t, err)

	db.TemplateTestClientSessionCRUD(t, sessionDB)
}
