package modelparam

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistryReplacePublishesSnapshot(t *testing.T) {
	first, err := Compile(nil)
	require.NoError(t, err)
	second, err := Compile([]Rule{{Enabled: true, Model: ModelSelector{Match: MatchExact, Pattern: "x"}, Path: "$.x", Constraint: Constraint{Operator: OperatorRequired}}})
	require.NoError(t, err)
	registry := NewRegistry(first)
	require.Same(t, first, registry.Snapshot())
	registry.Replace(second)
	require.Same(t, second, registry.Snapshot())
}
