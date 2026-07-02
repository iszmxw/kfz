package models

import "time"

type Book struct {
	Isbn        string    `gorm:"primaryKey;column:isbn" json:"isbn"`     // 归一化后的 ISBN
	Title       string    `gorm:"column:title" json:"title"`              // 书名
	Author      string    `gorm:"column:author" json:"author"`            // 作者
	Publisher   string    `gorm:"column:publisher" json:"publisher"`      // 出版社
	PublishYear string    `gorm:"column:publish_year" json:"publishYear"` // 出版年份
	CoverURL    string    `gorm:"column:cover_url" json:"coverUrl"`       // 封面地址
	Source      string    `gorm:"column:source" json:"source"`            // 数据来源
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`     // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`     // 更新时间
}

func (m *Book) TableName() string {
	return tableName("book")
}

var BookColumns = struct {
	Isbn        string
	Title       string
	Author      string
	Publisher   string
	PublishYear string
	CoverURL    string
	Source      string
	CreatedAt   string
	UpdatedAt   string
}{
	Isbn:        "isbn",
	Title:       "title",
	Author:      "author",
	Publisher:   "publisher",
	PublishYear: "publish_year",
	CoverURL:    "cover_url",
	Source:      "source",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}
