package postgres_adapter

import (
	"testing"

	"github.com/gianlucafrei/GoAM/internal/db"

	"github.com/stretchr/testify/require"
)

func TestPostgresRealmDB(t *testing.T) {
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close()

	realmDB, err := NewPostgresRealmDB(conn)
	require.NoError(t, err)

	db.TemplateTestRealmCRUD(t, realmDB)
}
