package service

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

func TestImageBillingSettlesActualCountUsingPriceSnapshot(t *testing.T) {
	ctx, _ := gin.CreateTestContext(nil)
	for _, tc := range []struct{ count, quota int }{{2, 25000}, {1, 12500}, {0, 0}} {
		info := &relaycommon.RelayInfo{OriginModelName: "image-test", PriceData: hosttypes.PriceData{
			UsePrice: true, ModelPrice: 0.05,
			ImageBilling: &hosttypes.ImageBillingDetail{BasePrice: 0.1, Multiplier: 0.5, GroupRatio: 0.5, QuotaPerUnit: 500000, Count: tc.count},
		}}
		info.PriceData.AddOtherRatio("n", 10) // Requested count must not override actual zero or one.
		summary := calculateTextQuotaSummary(ctx, info, &dto.Usage{PromptTokens: 1, TotalTokens: 1})
		assert.Equal(t, tc.quota, summary.Quota)
		require.Nil(t, info.QuotaClamp)
	}
}

func TestImageBillingSaturationIsAudited(t *testing.T) {
	ctx, _ := gin.CreateTestContext(nil)
	info := &relaycommon.RelayInfo{PriceData: hosttypes.PriceData{UsePrice: true,
		ImageBilling: &hosttypes.ImageBillingDetail{BasePrice: 1e20, Multiplier: 1, GroupRatio: 1, QuotaPerUnit: 500000, Count: 128},
	}}
	summary := calculateTextQuotaSummary(ctx, info, &dto.Usage{PromptTokens: 1})
	assert.Equal(t, common.MaxQuota, summary.Quota)
	require.NotNil(t, info.QuotaClamp)
}
