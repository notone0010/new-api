package modelparam

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/stretchr/testify/require"
)

func TestToAPIErrorReturnsNonRetryableBadRequest(t *testing.T) {
	err := ToAPIError(&Violation{Parameter: "temperature", Code: "model_parameter_invalid", Message: "invalid temperature"})
	require.Equal(t, http.StatusBadRequest, err.StatusCode)
	require.Equal(t, types.ErrorCode("model_parameter_invalid"), err.GetErrorCode())
	require.True(t, err.ToOpenAIError().Param == "temperature")
}
