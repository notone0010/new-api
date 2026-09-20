package controller

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/types"
	"github.com/shopspring/decimal"
)

type billingDetail struct {
	Name, Unit, Mode, Size, Tier, Note                      string
	Quantity, BasePrice, Discount, GroupRatio, Price, Quota decimal.Decimal
}

// billingDetails uses only the price snapshot persisted on the consume log.
// Unsupported or incomplete snapshots remain an explicitly unsplit settled charge.
func billingDetails(row model.BillingExportRow, quotaPerUnit decimal.Decimal) []billingDetail {
	actual := decimal.NewFromInt(row.Quota)
	fallback := []billingDetail{{Name: "请求结算", Unit: "次", Mode: "实际结算", Quantity: decimal.NewFromInt(1), Quota: actual, Note: "历史价格信息不足或计费规则无法可靠拆分，金额以实际扣费为准"}}
	var snapshot struct {
		Image              *types.ImageBillingDetail `json:"image_billing"`
		Mode               string                    `json:"billing_mode"`
		ModelRatio         *decimal.Decimal          `json:"model_ratio"`
		CompletionRatio    *decimal.Decimal          `json:"completion_ratio"`
		GroupRatio         *decimal.Decimal          `json:"group_ratio"`
		ModelPrice         *decimal.Decimal          `json:"model_price"`
		CacheTokens        int64                     `json:"cache_tokens"`
		CacheRatio         *decimal.Decimal          `json:"cache_ratio"`
		Semantic           string                    `json:"usage_semantic"`
		CacheCreation      int64                     `json:"cache_creation_tokens"`
		CacheCreationRatio *decimal.Decimal          `json:"cache_creation_ratio"`
		Cache5m            int64                     `json:"cache_creation_tokens_5m"`
		Cache1h            int64                     `json:"cache_creation_tokens_1h"`
		Cache5mRatio       *decimal.Decimal          `json:"cache_creation_ratio_5m"`
		Cache1hRatio       *decimal.Decimal          `json:"cache_creation_ratio_1h"`
	}
	if common.UnmarshalJsonStr(row.Other, &snapshot) != nil {
		return fallback
	}
	if image := snapshot.Image; image != nil {
		if image.Count < 0 || image.BasePrice < 0 || image.Multiplier < 0 || image.GroupRatio < 0 {
			return fallback
		}
		tier := "默认价格"
		if image.Tier > 0 || image.MaxPixels != nil || image.MinPixels > 0 {
			tier = fmt.Sprintf("像素 > %d", image.MinPixels)
			if image.MaxPixels != nil {
				tier += fmt.Sprintf(" 且 ≤ %d", *image.MaxPixels)
			}
		}
		base, discount, group := decimal.NewFromFloat(image.BasePrice), decimal.NewFromFloat(image.Multiplier), decimal.NewFromFloat(image.GroupRatio)
		return []billingDetail{{Name: "图片生成", Unit: "张", Mode: "按图", Size: image.Size, Tier: tier, Quantity: decimal.NewFromInt(int64(image.Count)), BasePrice: base, Discount: discount, GroupRatio: group, Price: base.Mul(discount).Mul(group), Quota: actual}}
	}
	if snapshot.Mode == "tiered_expr" {
		fallback[0].Mode = "按表达式"
		fallback[0].Note = "表达式整体结算；不将非线性表达式推算为固定单价"
		return fallback
	}
	if snapshot.GroupRatio == nil || snapshot.GroupRatio.IsNegative() {
		return fallback
	}
	group := *snapshot.GroupRatio
	if snapshot.ModelPrice != nil && !snapshot.ModelPrice.IsNegative() {
		price := snapshot.ModelPrice.Mul(group)
		if price.Mul(quotaPerUnit).Sub(actual).Abs().GreaterThan(decimal.NewFromInt(1)) {
			return fallback
		}
		return []billingDetail{{Name: "请求调用", Unit: "次", Mode: "按次", Quantity: decimal.NewFromInt(1), BasePrice: *snapshot.ModelPrice, Discount: decimal.NewFromInt(1), GroupRatio: group, Price: price, Quota: actual}}
	}
	if snapshot.ModelRatio == nil || snapshot.CompletionRatio == nil || snapshot.ModelRatio.IsNegative() || snapshot.CompletionRatio.IsNegative() {
		return fallback
	}
	prompt := row.PromptTokens
	creation := snapshot.CacheCreation
	if creation < 0 || snapshot.Cache5m < 0 || snapshot.Cache1h < 0 {
		return fallback
	}
	if snapshot.Semantic != "anthropic" {
		prompt -= snapshot.CacheTokens
		prompt -= creation
		snapshot.Cache5m, snapshot.Cache1h = 0, 0
	} else {
		creation = max(0, creation-snapshot.Cache5m-snapshot.Cache1h)
	}
	if prompt < 0 || row.CompletionTokens < 0 || snapshot.CacheTokens < 0 {
		return fallback
	}
	base := snapshot.ModelRatio.Mul(decimal.NewFromInt(1000)).Div(quotaPerUnit)
	parts := []struct {
		name  string
		count int64
		ratio *decimal.Decimal
	}{
		{"推理（输入）", prompt, nil}, {"推理（输出）", row.CompletionTokens, snapshot.CompletionRatio}, {"推理（缓存命中）", snapshot.CacheTokens, snapshot.CacheRatio},
		{"推理（缓存创建）", creation, snapshot.CacheCreationRatio}, {"推理（缓存创建5分钟）", snapshot.Cache5m, snapshot.Cache5mRatio}, {"推理（缓存创建1小时）", snapshot.Cache1h, snapshot.Cache1hRatio},
	}
	var details []billingDetail
	total := decimal.Zero
	for _, part := range parts {
		if part.count == 0 {
			continue
		}
		if part.name != "推理（输入）" && part.ratio == nil {
			return fallback
		}
		price := base
		if part.ratio != nil {
			if part.ratio.IsNegative() {
				return fallback
			}
			price = price.Mul(*part.ratio)
		}
		quantity := decimal.NewFromInt(part.count).Div(decimal.NewFromInt(1000))
		quota := quantity.Mul(price).Mul(group).Mul(quotaPerUnit)
		details = append(details, billingDetail{Name: part.name, Unit: "千tokens", Mode: "按Token", Quantity: quantity, BasePrice: price, Discount: decimal.NewFromInt(1), GroupRatio: group, Price: price.Mul(group), Quota: quota})
		total = total.Add(quota)
	}
	if len(details) == 0 || total.Sub(actual).Abs().GreaterThan(decimal.NewFromInt(1)) {
		return fallback
	}
	if !total.Equal(actual) {
		details = append(details, billingDetail{Name: "结算舍入差额", Mode: "结算调整", Quota: actual.Sub(total), Note: "请求结算取整产生的差额"})
	}
	return details
}
