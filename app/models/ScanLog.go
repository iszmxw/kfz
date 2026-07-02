package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type ScanLog struct {
	ID                    string           `gorm:"primaryKey;column:id" json:"id"`                                                 // 主键
	Isbn                  string           `gorm:"column:isbn" json:"isbn"`                                                        // 原始或归一化 ISBN
	NormalizedIsbn        string           `gorm:"column:normalized_isbn" json:"normalizedIsbn"`                                   // 归一化后的 ISBN
	BatchID               *string          `gorm:"column:batch_id" json:"batchId"`                                                 // 批次 ID
	StoreID               *string          `gorm:"column:store_id" json:"storeId"`                                                 // 门店 ID
	OperatorID            *string          `gorm:"column:operator_id" json:"operatorId"`                                           // 操作员 ID
	Decision              string           `gorm:"column:decision" json:"decision"`                                                // 判断结果
	Reason                string           `gorm:"column:reason" json:"reason"`                                                    // 判断原因
	MarketMinPrice        *decimal.Decimal `gorm:"column:market_min_price;type:decimal(20,2)" json:"marketMinPrice"`               // 判断时最低价
	MarketAvgPrice        *decimal.Decimal `gorm:"column:market_avg_price;type:decimal(20,2)" json:"marketAvgPrice"`               // 判断时平均价
	MarketMaxPrice        *decimal.Decimal `gorm:"column:market_max_price;type:decimal(20,2)" json:"marketMaxPrice"`               // 判断时最高价
	MarketSampleCount     *int             `gorm:"column:market_sample_count" json:"marketSampleCount"`                            // 判断时样本数
	SuggestedRecyclePrice *decimal.Decimal `gorm:"column:suggested_recycle_price;type:decimal(20,2)" json:"suggestedRecyclePrice"` // 建议回收价
	Confidence            string           `gorm:"column:confidence" json:"confidence"`                                            // 数据可信度
	PriceSnapshotID       *string          `gorm:"column:price_snapshot_id" json:"priceSnapshotId"`                                // 使用的价格快照 ID
	DuplicateRecently     bool             `gorm:"column:duplicate_recently" json:"duplicateRecently"`                             // 是否近期扫过
	DuplicateInBatch      bool             `gorm:"column:duplicate_in_batch" json:"duplicateInBatch"`                              // 是否当前批次重复
	ClientRequestID       *string          `gorm:"column:client_request_id" json:"clientRequestId"`                                // 请求幂等 ID
	ScannedAt             time.Time        `gorm:"column:scanned_at" json:"scannedAt"`                                             // 扫码时间
	CreatedAt             time.Time        `gorm:"column:created_at" json:"createdAt"`                                             // 创建时间
}

func (m *ScanLog) TableName() string {
	return tableName("scan_log")
}

var ScanLogColumns = struct {
	ID                    string
	Isbn                  string
	NormalizedIsbn        string
	BatchID               string
	StoreID               string
	OperatorID            string
	Decision              string
	Reason                string
	MarketMinPrice        string
	MarketAvgPrice        string
	MarketMaxPrice        string
	MarketSampleCount     string
	SuggestedRecyclePrice string
	Confidence            string
	PriceSnapshotID       string
	DuplicateRecently     string
	DuplicateInBatch      string
	ClientRequestID       string
	ScannedAt             string
	CreatedAt             string
}{
	ID:                    "id",
	Isbn:                  "isbn",
	NormalizedIsbn:        "normalized_isbn",
	BatchID:               "batch_id",
	StoreID:               "store_id",
	OperatorID:            "operator_id",
	Decision:              "decision",
	Reason:                "reason",
	MarketMinPrice:        "market_min_price",
	MarketAvgPrice:        "market_avg_price",
	MarketMaxPrice:        "market_max_price",
	MarketSampleCount:     "market_sample_count",
	SuggestedRecyclePrice: "suggested_recycle_price",
	Confidence:            "confidence",
	PriceSnapshotID:       "price_snapshot_id",
	DuplicateRecently:     "duplicate_recently",
	DuplicateInBatch:      "duplicate_in_batch",
	ClientRequestID:       "client_request_id",
	ScannedAt:             "scanned_at",
	CreatedAt:             "created_at",
}
