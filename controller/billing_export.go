package controller

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

const (
	billingCurrencyUSD = "USD"
	billingCurrencyCNY = "CNY"
)

type billingExportOptions struct {
	PeriodLabel  string
	Currency     string
	ExchangeRate decimal.Decimal
	QuotaPerUnit decimal.Decimal
}

type billingExportSummary struct {
	TotalQuota decimal.Decimal
}

func parseBillingPeriod(startDate, endDate string) (time.Time, time.Time, error) {
	startTime, err := time.ParseInLocation(time.DateOnly, startDate, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date: %w", err)
	}
	endInclusive, err := time.ParseInLocation(time.DateOnly, endDate, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date: %w", err)
	}
	if endInclusive.Before(startTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("end_date must not be before start_date")
	}
	return startTime, endInclusive.AddDate(0, 0, 1), nil
}

func safeSpreadsheetCell(value string) string {
	trimmed := strings.TrimLeftFunc(value, unicode.IsSpace)
	if trimmed == "" || !strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return value
	}
	return "'" + value
}

func buildBillingCSV(rows []model.BillingExportRow, options billingExportOptions) ([]byte, billingExportSummary, error) {
	var buffer bytes.Buffer
	buffer.Write([]byte{0xef, 0xbb, 0xbf})
	writer := csv.NewWriter(&buffer)
	header := []string{
		"账务账期/账单信息", "产品/产品信息", "配置名称/产品信息", "实例名称/资源信息", "计费单元/资源信息", "计费方式/资源信息",
		"项目备注/分账信息", "标签/分账信息", "标签备注/分账信息", "使用时长/时长信息", "时长单位/时长信息",
		"用量/用量信息", "用量单位/用量信息", "单价价格类型/价格信息", "单价/价格信息", "单价区间/价格信息", "单价单位/价格信息", "币种/原始费用金额", "原价/原始费用金额",
		"请求ID", "请求时间", "日志ID", "客户ID", "客户名称", "API令牌ID", "API令牌名称", "请求尺寸", "基础单价", "像素折扣", "分组倍率", "结算额度", "原始金额（美元）", "汇率", "备注",
	}
	if err := writer.Write(header); err != nil {
		return nil, billingExportSummary{}, err
	}

	summary := billingExportSummary{}
	for _, row := range rows {
		tokenName := row.TokenName
		if tokenName == "" {
			tokenName = fmt.Sprintf("Deleted token (%d)", row.TokenID)
		}
		quotaPerUnit := options.QuotaPerUnit
		var snapshot struct {
			Image struct {
				QuotaPerUnit float64 `json:"quota_per_unit"`
			} `json:"image_billing"`
		}
		if common.UnmarshalJsonStr(row.Other, &snapshot) == nil && snapshot.Image.QuotaPerUnit > 0 && !math.IsInf(snapshot.Image.QuotaPerUnit, 0) {
			quotaPerUnit = decimal.NewFromFloat(snapshot.Image.QuotaPerUnit)
		}
		details := billingDetails(row, quotaPerUnit)
		remainingUSD := decimal.NewFromInt(row.Quota).Div(quotaPerUnit).Round(6)
		remainingAmount := decimal.NewFromInt(row.Quota).Div(quotaPerUnit).Mul(options.ExchangeRate).Round(6)
		for i, detail := range details {
			usd := detail.Quota.Div(quotaPerUnit).Round(6)
			amount := detail.Quota.Div(quotaPerUnit).Mul(options.ExchangeRate).Round(6)
			if i == len(details)-1 {
				usd, amount = remainingUSD, remainingAmount
			}
			remainingUSD, remainingAmount = remainingUSD.Sub(usd), remainingAmount.Sub(amount)
			price, base, discount, group := "-", "-", "-", "-"
			priceType := "实际结算"
			if detail.Note == "" {
				price, base = detail.Price.Mul(options.ExchangeRate).String(), detail.BasePrice.Mul(options.ExchangeRate).String()
				discount, group = detail.Discount.String(), detail.GroupRatio.String()
				priceType = "固定单价"
				if detail.Mode == "按图" {
					priceType = "像素阶梯单价"
				}
			}
			record := []string{time.Unix(row.CreatedAt, 0).In(time.Local).Format("2006-01"), "AI模型", row.ModelName, tokenName, row.ModelName + "-" + detail.Name, detail.Mode,
				"-", "-", "-", "-", "-", detail.Quantity.String(), detail.Unit, priceType, price, detail.Tier, detail.Unit, options.Currency, amount.StringFixed(6),
				row.RequestID, time.Unix(row.CreatedAt, 0).In(time.Local).Format(time.RFC3339), strconv.Itoa(row.ID), strconv.Itoa(row.UserID), row.Username, strconv.Itoa(row.TokenID), tokenName,
				detail.Size, base, discount, group, detail.Quota.String(), usd.StringFixed(6), options.ExchangeRate.String(), detail.Note}
			// Only free-text fields need formula escaping; signed numeric adjustments remain numeric.
			for _, index := range []int{2, 3, 4, 15, 19, 23, 25, 26, 33} {
				record[index] = safeSpreadsheetCell(record[index])
			}
			if err := writer.Write(record); err != nil {
				return nil, billingExportSummary{}, err
			}
		}
		summary.TotalQuota = summary.TotalQuota.Add(decimal.NewFromInt(row.Quota))
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, billingExportSummary{}, err
	}
	return buffer.Bytes(), summary, nil
}

func ExportBillingCSV(c *gin.Context) {
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil || userID <= 0 {
		billingExportError(c, http.StatusBadRequest, "invalid user_id")
		return
	}
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	startTime, endTime, err := parseBillingPeriod(startDate, endDate)
	if err != nil {
		billingExportError(c, http.StatusBadRequest, err.Error())
		return
	}
	startTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()
	currency := strings.ToUpper(c.Query("currency"))
	if currency != billingCurrencyUSD && currency != billingCurrencyCNY {
		billingExportError(c, http.StatusBadRequest, "invalid currency")
		return
	}
	exchangeRate, err := decimal.NewFromString(c.Query("exchange_rate"))
	if err != nil || exchangeRate.LessThanOrEqual(decimal.Zero) || (currency == billingCurrencyUSD && !exchangeRate.Equal(decimal.NewFromInt(1))) {
		billingExportError(c, http.StatusBadRequest, "invalid exchange_rate")
		return
	}
	if math.IsNaN(common.QuotaPerUnit) || math.IsInf(common.QuotaPerUnit, 0) || common.QuotaPerUnit <= 0 {
		billingExportError(c, http.StatusInternalServerError, "invalid QuotaPerUnit configuration")
		return
	}
	quotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
	user, err := model.GetUserById(userID, false)
	if err != nil || user == nil {
		billingExportError(c, http.StatusNotFound, "user not found")
		return
	}

	rows, err := model.GetBillingExportRows(c.Request.Context(), model.BillingExportFilter{UserID: userID, StartTimestamp: startTimestamp, EndTimestamp: endTimestamp})
	if err != nil {
		billingExportError(c, http.StatusInternalServerError, "failed to query billing data")
		return
	}
	if len(rows) == 0 {
		billingExportError(c, http.StatusNotFound, "no billable usage found")
		return
	}
	periodLabel := fmt.Sprintf("%s/%s", startDate, endDate)
	data, summary, err := buildBillingCSV(rows, billingExportOptions{PeriodLabel: periodLabel, Currency: currency, ExchangeRate: exchangeRate, QuotaPerUnit: quotaPerUnit})
	if err != nil {
		billingExportError(c, http.StatusInternalServerError, "failed to build billing export")
		return
	}

	filename := fmt.Sprintf("billing-user-%d-%s-%s-%s.csv", userID, startDate, endDate, currency)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Header("Content-Length", strconv.Itoa(len(data)))
	c.Status(http.StatusOK)
	if _, err := c.Writer.Write(data); err != nil {
		common.SysError("failed to write billing export: " + err.Error())
		return
	}
	recordManageAuditFor(c, userID, "billing.export", map[string]interface{}{
		"target_user_id": userID, "start_date": startDate, "end_date": endDate,
		"currency": currency, "exchange_rate": exchangeRate.String(), "row_count": len(rows), "total_quota": summary.TotalQuota.String(),
	})
}

func billingExportError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"success": false, "message": message})
}
