package model

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupBillingExportLogDB(t *testing.T) {
	t.Helper()
	oldLogDB := LOG_DB
	t.Cleanup(func() { LOG_DB = oldLogDB })

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Log{}))
	LOG_DB = db
}

func createBillingExportLog(t *testing.T, log Log) {
	t.Helper()
	require.NoError(t, LOG_DB.Create(&log).Error)
}

func TestGetBillingExportRowsPreservesIndividualRequests(t *testing.T) {
	setupBillingExportLogDB(t)
	for _, log := range []Log{
		{UserId: 1, Username: "alice", TokenId: 11, TokenName: "project-a", ModelName: "model-a", Type: LogTypeConsume, CreatedAt: 1100, PromptTokens: 10, CompletionTokens: 5, Quota: 200},
		{UserId: 1, Username: "alice", TokenId: 11, TokenName: "project-a", ModelName: "model-a", Type: LogTypeConsume, CreatedAt: 1200, PromptTokens: 20, CompletionTokens: 7, Quota: 220},
		{UserId: 1, Username: "alice", TokenId: 12, TokenName: "project-b", ModelName: "model-a", Type: LogTypeConsume, CreatedAt: 1300, PromptTokens: 3, CompletionTokens: 4, Quota: 70},
		{UserId: 1, Username: "alice", TokenId: 11, TokenName: "project-a", ModelName: "model-b", Type: LogTypeConsume, CreatedAt: 1400, PromptTokens: 8, CompletionTokens: 9, Quota: 170},
	} {
		createBillingExportLog(t, log)
	}

	rows, err := GetBillingExportRows(context.Background(), BillingExportFilter{UserID: 1, StartTimestamp: 1000, EndTimestamp: 2000})

	require.NoError(t, err)
	require.Len(t, rows, 4)
	assert.Equal(t, 1, rows[0].RequestCount)
	assert.Equal(t, int64(10), rows[0].PromptTokens)
	assert.Equal(t, int64(5), rows[0].CompletionTokens)
	assert.Equal(t, int64(200), rows[0].Quota)
	assert.Equal(t, 11, rows[0].TokenID)
	assert.Equal(t, "model-a", rows[0].ModelName)
}

func TestGetBillingExportRowsUsesUserConsumeAndHalfOpenPeriod(t *testing.T) {
	setupBillingExportLogDB(t)
	for _, log := range []Log{
		{UserId: 1, Type: LogTypeConsume, CreatedAt: 999, Quota: 1},
		{UserId: 1, Type: LogTypeConsume, CreatedAt: 1000, Quota: 10},
		{UserId: 1, Type: LogTypeRefund, CreatedAt: 1500, Quota: 100},
		{UserId: 2, Type: LogTypeConsume, CreatedAt: 1500, Quota: 1000},
		{UserId: 1, Type: LogTypeConsume, CreatedAt: 1999, Quota: 20},
		{UserId: 1, Type: LogTypeConsume, CreatedAt: 2000, Quota: 2},
	} {
		createBillingExportLog(t, log)
	}

	rows, err := GetBillingExportRows(context.Background(), BillingExportFilter{UserID: 1, StartTimestamp: 1000, EndTimestamp: 2000})

	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, int64(10), rows[0].Quota)
	assert.Equal(t, int64(20), rows[1].Quota)
}

func TestGetBillingExportRowsPreservesStoredDeletedTokenIdentity(t *testing.T) {
	setupBillingExportLogDB(t)
	createBillingExportLog(t, Log{UserId: 1, TokenId: 99, TokenName: "", ModelName: "model-a", Type: LogTypeConsume, CreatedAt: 1100, Quota: 10})

	rows, err := GetBillingExportRows(context.Background(), BillingExportFilter{UserID: 1, StartTimestamp: 1000, EndTimestamp: 2000})

	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, 99, rows[0].TokenID)
	assert.Empty(t, rows[0].TokenName)
}
