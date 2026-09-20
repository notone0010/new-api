package modelparam

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigUsesCommonJSONWrapperAndCompilesRules(t *testing.T) {
	snapshot, err := LoadConfig([]byte(`{"version":1,"rules":[{"enabled":true,"model":{"match":"exact","pattern":"gpt-test"},"protocols":["openai_chat"],"path":"$.temperature","constraint":{"operator":"range","min":"0","max":"1"}}]}`))
	require.NoError(t, err)
	require.NotNil(t, snapshot)
}

func TestConfigureDefaultKeepsPreviousSnapshotOnInvalidConfig(t *testing.T) {
	SetDefaultRegistry(nil)
	require.NoError(t, ConfigureDefault([]byte(`{"version":1,"rules":[]}`)))
	previous := DefaultRegistry().Snapshot()
	require.Error(t, ConfigureDefault([]byte(`{"version":1,"rules":[{"enabled":true,"model":{"match":"exact","pattern":"x"},"path":"$..bad","constraint":{"operator":"required"}}]}`)))
	require.Same(t, previous, DefaultRegistry().Snapshot())
}
