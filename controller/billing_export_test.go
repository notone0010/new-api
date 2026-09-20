package controller

import (
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildBillingCSVUsesBOMAndReconcilesQuota(t *testing.T) {
	data, summary, err := buildBillingCSV([]model.BillingExportRow{{
		UserID: 1, Username: "alice", TokenID: 11, TokenName: "project-a", ModelName: "gpt-test",
		RequestCount: 2, PromptTokens: 1000, CompletionTokens: 200, Quota: 750000,
	}}, billingExportOptions{
		PeriodLabel: "2026-09-01/2026-10-01", Currency: billingCurrencyCNY,
		ExchangeRate: decimal.RequireFromString("7.2"), QuotaPerUnit: decimal.NewFromInt(500000),
	})

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(data), 3)
	assert.Equal(t, []byte{0xef, 0xbb, 0xbf}, data[:3])
	assert.True(t, decimal.NewFromInt(750000).Equal(summary.TotalQuota))
	records, err := csv.NewReader(strings.NewReader(string(data[3:]))).ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, "账务账期/账单信息", records[0][0])
	assert.Equal(t, "请求ID", records[0][19])
	assert.Equal(t, "1.500000", records[1][31])
	assert.Equal(t, "10.800000", records[1][18])
}

func TestParseBillingPeriodUsesServerLocalCalendarDays(t *testing.T) {
	previousLocal := time.Local
	time.Local = time.FixedZone("UTC+8", 8*60*60)
	t.Cleanup(func() { time.Local = previousLocal })

	start, end, err := parseBillingPeriod("2026-09-01", "2026-09-30")

	require.NoError(t, err)
	assert.Equal(t, "2026-09-01T00:00:00+08:00", start.Format(time.RFC3339))
	assert.Equal(t, "2026-10-01T00:00:00+08:00", end.Format(time.RFC3339))
}

func TestParseBillingPeriodRejectsInvalidRange(t *testing.T) {
	_, _, err := parseBillingPeriod("2026-10-01", "2026-09-30")

	require.ErrorContains(t, err, "end_date")
}

func TestExportBillingCSVRejectsInvalidCalendarRangeBeforeDatabaseAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/log/billing/export?user_id=1&start_date=2026-10-01&end_date=2026-09-30&currency=CNY&exchange_rate=7.2", nil)

	ExportBillingCSV(context)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "end_date")
}

func TestBuildBillingCSVNeutralizesSpreadsheetFormulaCells(t *testing.T) {
	data, _, err := buildBillingCSV([]model.BillingExportRow{{
		Username: "=cmd", TokenID: 9, TokenName: "+token", ModelName: "@model",
	}}, billingExportOptions{
		PeriodLabel: "period", Currency: billingCurrencyUSD,
		ExchangeRate: decimal.NewFromInt(1), QuotaPerUnit: decimal.NewFromInt(500000),
	})

	require.NoError(t, err)
	assert.Contains(t, string(data), "'=cmd")
	assert.Contains(t, string(data), "'+token")
	assert.Contains(t, string(data), "'@model")
}

func TestBuildBillingCSVRequestDetailsAmountsReconcile(t *testing.T) {
	rows := []model.BillingExportRow{
		{ID: 1, RequestID: "req-1", PromptTokens: 3, CompletionTokens: 1, Quota: 5, Other: `{"model_ratio":1,"completion_ratio":1.6,"group_ratio":1,"model_price":-1}`},
		{ID: 2, RequestID: "req-2", Quota: 25000, Other: `{"image_billing":{"base_price":0.1,"multiplier":0.5,"group_ratio":1,"count":1,"size":"1024x1024","quota_per_unit":500000}}`},
	}
	data, _, err := buildBillingCSV(rows, billingExportOptions{Currency: billingCurrencyCNY, ExchangeRate: decimal.RequireFromString("7.3"), QuotaPerUnit: decimal.NewFromInt(500000)})
	require.NoError(t, err)
	records, err := csv.NewReader(strings.NewReader(string(data[3:]))).ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 5)
	sums := map[string]decimal.Decimal{}
	for _, record := range records[1:] {
		amount, err := decimal.NewFromString(record[18])
		require.NoError(t, err)
		sums[record[19]] = sums[record[19]].Add(amount)
	}
	assert.Equal(t, "0.000073", sums["req-1"].String())
	assert.Equal(t, "0.365", sums["req-2"].String())
}
