package modelparam

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompilePath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantGJSON string
		wantWild  bool
	}{
		{name: "field", path: "$.temperature", wantGJSON: "temperature"},
		{name: "nested", path: "$.generationConfig.maxOutputTokens", wantGJSON: "generationConfig.maxOutputTokens"},
		{name: "index", path: "$.tools[0].type", wantGJSON: "tools.0.type"},
		{name: "wildcard", path: "$.tools[*].type", wantGJSON: "tools.#.type", wantWild: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled, err := CompilePath(tt.path)
			require.NoError(t, err)
			require.Equal(t, tt.wantGJSON, compiled.GJSON)
			require.Equal(t, tt.wantWild, compiled.HasWildcard)
		})
	}
}
