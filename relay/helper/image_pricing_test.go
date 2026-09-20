package helper

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	kittypes "github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImagePricingPreconsumeSnapshotAndMissingRequestDimensions(t *testing.T) {
	previous := ratio_setting.ImagePrice2JSONString()
	groups := ratio_setting.GroupRatio2JSONString()
	quotaUnit := common.QuotaPerUnit
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateImagePriceByJSONString(previous))
		require.NoError(t, ratio_setting.UpdateGroupRatioByJSONString(groups))
		common.QuotaPerUnit = quotaUnit
	})
	common.QuotaPerUnit = 500000
	require.NoError(t, ratio_setting.UpdateGroupRatioByJSONString(`{"image-test-group":0.5}`))
	require.NoError(t, ratio_setting.UpdateImagePriceByJSONString(`{"dall-e-3":{"base_price":0.1,"pixel_tiers":[{"max_pixels":1048576,"multiplier":0.5},{"max_pixels":null,"multiplier":1}]}}`))
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("POST", "/v1/images/generations", strings.NewReader(`{"model":"dall-e-3","prompt":"test"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	_, err := GetAndValidOpenAIImageRequest(ctx, relayconstant.RelayModeImagesGenerations)
	require.ErrorContains(t, err, "explicit request size")
	n := uint(2)
	info := &relaycommon.RelayInfo{OriginModelName: "dall-e-3", UserGroup: "image-test-group", UsingGroup: "image-test-group", Request: &dto.ImageRequest{Size: "1024x1024", N: &n}}
	price, err := ModelPriceHelper(ctx, info, 0, &kittypes.TokenCountMeta{ImagePriceRatio: 3, BillingRatios: map[string]float64{"n": 2}})
	require.NoError(t, err)
	assert.Equal(t, 25000, price.QuotaToPreConsume)
	require.NotNil(t, price.ImageBilling)
	require.NoError(t, ratio_setting.UpdateImagePriceByJSONString(`{"dall-e-3":{"base_price":9}}`))
	assert.Equal(t, 0.1, price.ImageBilling.BasePrice)
	assert.Equal(t, 0.5, price.ImageBilling.Multiplier)
}

func TestImagePixelTiersUseTotalPixelsAndInclusiveBounds(t *testing.T) {
	var cfg ratio_setting.ImagePricingConfig
	require.NoError(t, common.UnmarshalJsonStr(`{"base_price":0.1,"pixel_tiers":[{"max_pixels":1048576,"multiplier":0.5},{"max_pixels":4194304,"multiplier":0.8},{"max_pixels":null,"multiplier":1.5}]}`, &cfg))
	require.NoError(t, cfg.Validate())
	var dimensionsRequest dto.ImageRequest
	require.NoError(t, common.UnmarshalJsonStr(`{"width":2048,"height":512}`, &dimensionsRequest))
	dimensions, err := ResolveImagePricing(&dimensionsRequest, cfg)
	require.NoError(t, err)
	assert.Equal(t, int64(1048576), dimensions.Pixels)
	assert.Equal(t, 0.5, dimensions.Multiplier)
	for _, tc := range []struct {
		size       string
		tier       int
		multiplier float64
	}{
		{"1024x1024", 1, 0.5}, {"2048x512", 1, 0.5}, {"1025x1024", 2, 0.8}, {"2048x2048", 2, 0.8}, {"2049x2048", 3, 1.5},
	} {
		t.Run(tc.size, func(t *testing.T) {
			detail, err := ResolveImagePricing(&dto.ImageRequest{Size: tc.size}, cfg)
			require.NoError(t, err)
			assert.Equal(t, tc.tier, detail.Tier)
			assert.Equal(t, tc.multiplier, detail.Multiplier)
			assert.InDelta(t, 0.1*tc.multiplier, detail.UnitPrice, 1e-12)
		})
	}
	for _, size := range []string{"", "auto", "1K", "1024x0", "-1x1024", "9223372036854775807x2", "1024.5x1024"} {
		t.Run("reject_"+size, func(t *testing.T) {
			_, err := ResolveImagePricing(&dto.ImageRequest{Size: size}, cfg)
			require.Error(t, err)
		})
	}
}

func TestImagePricingValidationDoesNotReplaceValidSettings(t *testing.T) {
	previous := ratio_setting.ImagePrice2JSONString()
	t.Cleanup(func() { require.NoError(t, ratio_setting.UpdateImagePriceByJSONString(previous)) })
	valid := `{"image-test":{"base_price":0.1,"pixel_tiers":[{"max_pixels":1048576,"multiplier":0.5},{"max_pixels":null,"multiplier":1}]}}`
	require.NoError(t, ratio_setting.UpdateImagePriceByJSONString(valid))
	for _, invalid := range []string{
		`{"base_price":-1}`, `{"pixel_tiers":[]}`, `{"base_price":1,"pixel_tiers":[{"max_pixels":100,"multiplier":1}]}`,
		`{"base_price":1,"pixel_tiers":[{"max_pixels":100,"multiplier":1},{"max_pixels":99,"multiplier":1},{"max_pixels":null,"multiplier":1}]}`,
		`{"base_price":1,"pixel_tiers":[{"max_pixels":null,"multiplier":0}]}`,
	} {
		require.Error(t, ratio_setting.UpdateImagePriceByJSONString(`{"image-test":`+invalid+`}`))
		cfg, ok := ratio_setting.GetImagePricingConfig("image-test")
		require.True(t, ok)
		assert.Equal(t, 0.1, *cfg.BasePrice)
	}
}
