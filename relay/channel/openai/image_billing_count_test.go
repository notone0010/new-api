package openai

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	hosttypes "github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageBillingUsesReportedCountIncludingZero(t *testing.T) {
	ctx, _ := gin.CreateTestContext(nil)
	for _, tc := range []struct {
		body   string
		count  int
		source string
	}{
		{`{"data":[{},{}],"usage":{"generated_images":1}}`, 1, "usage.generated_images"},
		{`{"data":[{}],"usage":{"generated_images":0}}`, 0, "usage.generated_images"},
		{`{"data":[{},{}]}`, 2, "data"},
	} {
		var response struct {
			Usage dto.Usage `json:"usage"`
		}
		require.NoError(t, common.UnmarshalJsonStr(tc.body, &response))
		info := &relaycommon.RelayInfo{PriceData: hosttypes.PriceData{UsePrice: true, ImageBilling: &hosttypes.ImageBillingDetail{Count: 3}}}
		updateOpenAIImageCountFromResponse(ctx, info, []byte(tc.body), &response.Usage)
		assert.Equal(t, tc.count, info.PriceData.ImageBilling.Count)
		assert.Equal(t, tc.source, info.PriceData.ImageBilling.CountSource)
	}
}
