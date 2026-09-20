package controller

import (
	"github.com/QuantumNous/new-api/model"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBillingDetailsSplitTokenSnapshotAndReconcile(t *testing.T) {
	row := model.BillingExportRow{PromptTokens: 1000, CompletionTokens: 100, Quota: 1260, Other: `{"model_ratio":1,"completion_ratio":3,"group_ratio":1,"model_price":-1,"cache_tokens":200,"cache_ratio":0.8}`}
	parts := billingDetails(row, decimal.NewFromInt(500000))
	require.Len(t, parts, 3)
	assert.Equal(t, "0.8", parts[0].Quantity.String())
	assert.Equal(t, "推理（缓存命中）", parts[2].Name)
	total := decimal.Zero
	for _, part := range parts {
		total = total.Add(part.Quota)
	}
	assert.Equal(t, "1260", total.String())
	row.Quota = 2000
	parts = billingDetails(row, decimal.NewFromInt(500000))
	require.Len(t, parts, 1)
	assert.NotEmpty(t, parts[0].Note)
	assert.Equal(t, "2000", parts[0].Quota.String())
}

func TestBillingDetailsImageUsesHistoricalSnapshot(t *testing.T) {
	row := model.BillingExportRow{Quota: 25000, Other: `{"image_billing":{"base_price":0.1,"multiplier":0.5,"group_ratio":0.5,"count":2,"size":"1024x1024","max_pixels":1048576}}`}
	parts := billingDetails(row, decimal.NewFromInt(500000))
	require.Len(t, parts, 1)
	assert.Equal(t, "2", parts[0].Quantity.String())
	assert.Equal(t, "0.025", parts[0].Price.String())
	assert.Equal(t, "1024x1024", parts[0].Size)
	assert.Contains(t, parts[0].Tier, "≤ 1048576")
	assert.Equal(t, "25000", parts[0].Quota.String())
}
