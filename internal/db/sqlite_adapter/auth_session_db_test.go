package sqlite_adapter

import (
	"testing"

	"github.com/Identityplane/GoAM/internal/db"

	"github.com/stretchr/testify/require"
)

func TestAuthSessionCRUD(t *testing.T) {
	sqldb := setupTestDB(t)
	authSessionDB, err := NewAuthSessionDB(sqldb)
	require.NoError(t, err)

	db.TemplateTestAuthSessionCRUD(t, authSessionDB)
}
