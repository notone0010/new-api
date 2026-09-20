package modelparam

import (
	"fmt"

	"github.com/QuantumNous/new-api/relaykit/types"
)

func ProtocolFor(format types.RelayFormat) (Protocol, error) {
	switch format {
	case types.RelayFormatOpenAI:
		return ProtocolOpenAIChat, nil
	case types.RelayFormatOpenAIResponses, types.RelayFormatOpenAIResponsesCompaction:
		return ProtocolOpenAIResponses, nil
	case types.RelayFormatOpenAIImage:
		return ProtocolOpenAIImage, nil
	case types.RelayFormatOpenAIAudio:
		return ProtocolOpenAIAudio, nil
	case types.RelayFormatEmbedding:
		return ProtocolOpenAIEmbedding, nil
	case types.RelayFormatRerank:
		return ProtocolRerank, nil
	case types.RelayFormatClaude:
		return ProtocolClaudeMessages, nil
	case types.RelayFormatGemini:
		return ProtocolGeminiGenerate, nil
	default:
		return "", fmt.Errorf("unsupported model parameter validation protocol: %s", format)
	}
}
