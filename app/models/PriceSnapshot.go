package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type PriceSnapshot struct {
	ID            uint64           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`        // 主键
	Isbn          string           `gorm:"column:isbn" json:"isbn"`                             // 关联 book.isbn
	Source        string           `gorm:"column:source" json:"source"`                         // 价格来源
	MinPrice      *decimal.Decimal `gorm:"column:min_price;type:decimal(20,2)" json:"minPrice"` // 最低价
	AvgPrice      *decimal.Decimal `gorm:"column:avg_price;type:decimal(20,2)" json:"avgPrice"` // 平均价
	MaxPrice      *decimal.Decimal `gorm:"column:max_price;type:decimal(20,2)" json:"maxPrice"` // 最高价
	SampleCount   int              `gorm:"column:sample_count" json:"sampleCount"`              // 有效样本数
	Confidence    string           `gorm:"column:confidence" json:"confidence"`                 // 数据可信度
	RawURL        string           `gorm:"column:raw_url" json:"rawUrl"`                        // 来源链接
	RawPayloadRef string           `gorm:"column:raw_payload_ref" json:"rawPayloadRef"`         // 原始数据存储引用
	CollectedAt   time.Time        `gorm:"column:collected_at" json:"collectedAt"`              // 采集或录入时间
	ExpiresAt     *time.Time       `gorm:"column:expires_at" json:"expiresAt"`                  // 过期时间
	CreatedAt     time.Time        `gorm:"column:created_at" json:"createdAt"`                  // 创建时间
}

func (m *PriceSnapshot) TableName() string {
	return tableName("price_snapshot")
}

var PriceSnapshotColumns = struct {
	ID            string
	Isbn          string
	Source        string
	MinPrice      string
	AvgPrice      string
	MaxPrice      string
	SampleCount   string
	Confidence    string
	RawURL        string
	RawPayloadRef string
	CollectedAt   string
	ExpiresAt     string
	CreatedAt     string
}{
	ID:            "id",
	Isbn:          "isbn",
	Source:        "source",
	MinPrice:      "min_price",
	AvgPrice:      "avg_price",
	MaxPrice:      "max_price",
	SampleCount:   "sample_count",
	Confidence:    "confidence",
	RawURL:        "raw_url",
	RawPayloadRef: "raw_payload_ref",
	CollectedAt:   "collected_at",
	ExpiresAt:     "expires_at",
	CreatedAt:     "created_at",
}
