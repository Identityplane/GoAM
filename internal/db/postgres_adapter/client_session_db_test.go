package postgres_adapter

import (
	"testing"

	"github.com/gianlucafrei/GoAM/internal/db"

	"github.com/stretchr/testify/require"
)

func TestPostgresClientSessionDB(t *testing.T) {
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close()

	sessionDB, err := NewPostgresClientSessionDB(conn)
	require.NoError(t, err)

	db.TemplateTestClientSessionCRUD(t, sessionDB)
}
