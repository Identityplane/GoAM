package sqlite_adapter

import (
	"testing"

	"github.com/gianlucafrei/GoAM/internal/db"

	"github.com/stretchr/testify/require"
)

func TestRealmCRUD(t *testing.T) {
	sqldb := setupTestDB(t)
	realmDB, err := NewRealmDB(sqldb)
	require.NoError(t, err)
	db.TemplateTestRealmCRUD(t, realmDB)
}
