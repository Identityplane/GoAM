package sqlite_adapter

import (
	"testing"

	"github.com/gianlucafrei/GoAM/internal/db"

	"github.com/stretchr/testify/require"
)

func TestSigningKeyCRUD(t *testing.T) {
	sqldb := setupTestDB(t)
	signingKeyDB, err := NewSigningKeyDB(sqldb)
	require.NoError(t, err)
	db.TemplateTestSigningKeyCRUD(t, signingKeyDB)
}
