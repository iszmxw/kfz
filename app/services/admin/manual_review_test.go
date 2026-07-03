package admin

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"goapi/app/models"
	"goapi/pkg/config"
	"goapi/pkg/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAdminServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	config.Add("database", config.StrMap{
		"mysql": map[string]interface{}{"prefix": ""},
	})
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Book{}, &models.ScanLog{}, &models.ManualDecision{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	original := mysql.DB
	mysql.DB = db
	t.Cleanup(func() {
		mysql.DB = original
	})
	return db
}

func TestCreateManualDecisionDoesNotMutateScanLog(t *testing.T) {
	db := setupAdminServiceTestDB(t)
	now := time.Now()
	book := models.Book{Isbn: "9787111128069", Title: "示例书", Source: "manual", CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("create book: %v", err)
	}
	scan := models.ScanLog{
		ID:             1,
		Isbn:           book.Isbn,
		NormalizedIsbn: book.Isbn,
		Decision:       "NEED_REVIEW",
		Reason:         "暂无有效价格数据，需要人工确认",
		Confidence:     "NONE",
		ScannedAt:      now,
		CreatedAt:      now,
	}
	if err := db.Create(&scan).Error; err != nil {
		t.Fatalf("create scan: %v", err)
	}
	price := decimal.RequireFromString("8.50")

	decision, err := CreateManualDecision(ManualDecisionInput{
		ScanLogID:          scan.ID,
		AdminUserID:        1,
		ManualDecision:     "ACCEPT",
		ActualRecyclePrice: &price,
		Note:               "人工确认可收",
	})
	if err != nil {
		t.Fatalf("CreateManualDecision err=%v", err)
	}
	if decision.ScanLogID != scan.ID {
		t.Fatalf("scan log id = %d, want %d", decision.ScanLogID, scan.ID)
	}

	var storedScan models.ScanLog
	if err := db.First(&storedScan, "id = ?", scan.ID).Error; err != nil {
		t.Fatalf("load scan: %v", err)
	}
	if storedScan.Decision != "NEED_REVIEW" {
		t.Fatalf("scan decision changed to %s, want NEED_REVIEW", storedScan.Decision)
	}
}
