package fuzz

import (
	"os"
	"testing"

	"github.com/gianlucafrei/GoAM/internal/lib"

	"github.com/stretchr/testify/require"
)

func FuzzConvertToFlowDefinitionLogic(f *testing.F) {
	// Seed with real YAML example
	content, err := os.ReadFile("../integration/config/tenants/acme/customers/flows/login_or_register.yaml")
	require.NoError(f, err)
	f.Add(string(content)) // seed with known valid YAML

	f.Fuzz(func(t *testing.T, input string) {

		_, _ = lib.LoadFlowDefinitonFromString(input)
	})
}
