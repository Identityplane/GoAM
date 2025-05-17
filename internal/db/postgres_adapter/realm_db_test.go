package postgres_adapter

import (
	"context"
	"testing"

	"goiam/internal/db"

	"github.com/stretchr/testify/require"
)

func TestPostgresRealmDB(t *testing.T) {
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close(context.Background())

	realmDB, err := NewPostgresRealmDB(conn)
	require.NoError(t, err)

	db.TemplateTestRealmCRUD(t, realmDB)
}
