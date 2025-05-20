package postgres_adapter

import (
	"testing"

	"goiam/internal/db"

	"github.com/stretchr/testify/require"
)

func TestPostgresApplicationDB(t *testing.T) {
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close()

	appDB, err := NewPostgresApplicationDB(conn)
	require.NoError(t, err)

	db.TemplateTestApplicationCRUD(t, appDB)
}
