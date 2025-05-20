package postgres_adapter

import (
	"testing"

	"goiam/internal/db"

	"github.com/stretchr/testify/require"
)

func TestPostgresAuthSessionDB(t *testing.T) {
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close()

	sessionDB, err := NewPostgresAuthSessionDB(conn)
	require.NoError(t, err)

	db.TemplateTestAuthSessionCRUD(t, sessionDB)
}
