package sqlite_adapter

import (
	"goiam/internal/db"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRealmCRUD(t *testing.T) {
	sqldb := setupTestDB(t)
	realmDB, err := NewRealmDB(sqldb)
	require.NoError(t, err)
	db.TemplateTestRealmCRUD(t, realmDB)
}
