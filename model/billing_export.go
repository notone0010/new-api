package model

import "context"

type BillingExportFilter struct {
	UserID         int
	StartTimestamp int64
	EndTimestamp   int64
}

type BillingExportRow struct {
	ID               int    `gorm:"column:id"`
	CreatedAt        int64  `gorm:"column:created_at"`
	RequestID        string `gorm:"column:request_id"`
	Other            string `gorm:"column:other"`
	UserID           int    `gorm:"column:user_id"`
	Username         string `gorm:"column:username"`
	TokenID          int    `gorm:"column:token_id"`
	TokenName        string `gorm:"column:token_name"`
	ModelName        string `gorm:"column:model_name"`
	RequestCount     int    `gorm:"column:request_count"`
	PromptTokens     int64  `gorm:"column:prompt_tokens"`
	CompletionTokens int64  `gorm:"column:completion_tokens"`
	Quota            int64  `gorm:"column:quota"`
}

func GetBillingExportRows(ctx context.Context, filter BillingExportFilter) ([]BillingExportRow, error) {
	rows := make([]BillingExportRow, 0)
	err := LOG_DB.WithContext(ctx).
		Table("logs").
		Select("id, created_at, request_id, other, user_id, username, token_id, token_name, model_name, 1 AS request_count, prompt_tokens, completion_tokens, quota").
		Where("user_id = ?", filter.UserID).
		Where("type = ?", LogTypeConsume).
		Where("created_at >= ?", filter.StartTimestamp).
		Where("created_at < ?", filter.EndTimestamp).
		Order("created_at ASC, id ASC").
		Find(&rows).Error
	return rows, err
}
