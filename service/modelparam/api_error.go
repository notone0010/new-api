package modelparam

import (
	"errors"
	"net/http"

	"github.com/QuantumNous/new-api/relaykit/types"
)

var ErrInvalidRequestJSON = errors.New("invalid JSON request body")

func ToAPIError(violation *Violation) *types.NewAPIError {
	if violation == nil {
		return nil
	}
	return types.WithOpenAIError(types.OpenAIError{
		Message: violation.Message,
		Type:    "invalid_request_error",
		Param:   violation.Parameter,
		Code:    violation.Code,
	}, http.StatusBadRequest, types.ErrOptionWithSkipRetry())
}

func ParseErrorToAPIError(err error) *types.NewAPIError {
	return types.WithOpenAIError(types.OpenAIError{
		Message: err.Error(),
		Type:    "invalid_request_error",
		Code:    types.ErrorCodeInvalidRequest,
	}, http.StatusBadRequest, types.ErrOptionWithSkipRetry())
}
