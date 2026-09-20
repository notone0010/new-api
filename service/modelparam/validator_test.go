package modelparam

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateRangeAndPreservesExplicitZero(t *testing.T) {
	snapshot, err := Compile([]Rule{{
		ID:        1,
		Enabled:   true,
		Model:     ModelSelector{Match: MatchExact, Pattern: "gpt-test"},
		Protocols: []Protocol{ProtocolOpenAIChat},
		Path:      "$.temperature",
		Constraint: Constraint{
			Operator: OperatorRange,
			Min:      "0",
			Max:      "1",
		},
		Error: ErrorDefinition{Code: "model_parameter_out_of_range", Message: "temperature must be between 0 and 1"},
	}})
	require.NoError(t, err)

	violation, err := snapshot.Validate(Input{
		Model:    "gpt-test",
		Protocol: ProtocolOpenAIChat,
		Body:     []byte(`{"temperature":0}`),
	})
	require.NoError(t, err)
	require.Nil(t, violation)
}

func TestValidateRejectsInvalidRangeWithHTTP400Metadata(t *testing.T) {
	snapshot, err := Compile([]Rule{{
		ID:        2,
		Enabled:   true,
		Model:     ModelSelector{Match: MatchGlob, Pattern: "gpt-*"},
		Protocols: []Protocol{ProtocolOpenAIChat},
		Path:      "$.temperature",
		Constraint: Constraint{
			Operator: OperatorRange,
			Min:      "0",
			Max:      "1",
		},
		Error: ErrorDefinition{Code: "model_parameter_out_of_range", Message: "temperature must be between 0 and 1"},
	}})
	require.NoError(t, err)

	violation, err := snapshot.Validate(Input{
		Model:    "gpt-5",
		Protocol: ProtocolOpenAIChat,
		Body:     []byte(`{"temperature":2}`),
	})
	require.NoError(t, err)
	require.NotNil(t, violation)
	require.Equal(t, "temperature", violation.Parameter)
	require.Equal(t, "model_parameter_out_of_range", violation.Code)
}

func TestCompileRejectsUnsupportedJSONPath(t *testing.T) {
	_, err := Compile([]Rule{{
		Enabled:    true,
		Model:      ModelSelector{Match: MatchExact, Pattern: "gpt-test"},
		Protocols:  []Protocol{ProtocolOpenAIChat},
		Path:       "$..temperature",
		Constraint: Constraint{Operator: OperatorRequired},
	}})
	require.Error(t, err)
}

func TestValidateWildcardRequiredRejectsMissingChild(t *testing.T) {
	snapshot, err := Compile([]Rule{{
		Enabled:    true,
		Model:      ModelSelector{Match: MatchExact, Pattern: "gpt-test"},
		Protocols:  []Protocol{ProtocolOpenAIChat},
		Path:       "$.tools[*].type",
		Constraint: Constraint{Operator: OperatorRequired},
		Error:      ErrorDefinition{Code: "required", Message: "tool type is required"},
	}})
	require.NoError(t, err)
	violation, err := snapshot.Validate(Input{Model: "gpt-test", Protocol: ProtocolOpenAIChat, Body: []byte(`{"tools":[{"type":"function"},{}]}`)})
	require.NoError(t, err)
	require.NotNil(t, violation)
	require.Equal(t, "tools[1].type", violation.Parameter)
}
