package sqlite_adapter

import (
	"testing"

	"github.com/gianlucafrei/GoAM/internal/db"

	"github.com/stretchr/testify/require"
)

func TestClientSessionCRUD(t *testing.T) {
	sqldb := setupTestDB(t)
	clientSessionDB, err := NewClientSessionDB(sqldb)
	require.NoError(t, err)

	db.TemplateTestClientSessionCRUD(t, clientSessionDB)
}
