package postgres_adapter

import (
	"testing"

	"github.com/gianlucafrei/GoAM/internal/db"

	"github.com/stretchr/testify/require"
)

func TestPostgresSigningKeyDB(t *testing.T) {
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close()

	signingKeyDB, err := NewPostgresSigningKeysDB(conn)
	require.NoError(t, err)

	db.TemplateTestSigningKeyCRUD(t, signingKeyDB)
}
