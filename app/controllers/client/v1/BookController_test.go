package v1

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

func setupControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	config.Add("database", config.StrMap{
		"mysql": map[string]interface{}{
			"prefix": "",
		},
	})

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Book{}, &models.PriceSnapshot{}, &models.ScanLog{}, &models.RecycleRule{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	originalDB := mysql.DB
	mysql.DB = db
	t.Cleanup(func() {
		mysql.DB = originalDB
	})

	return db
}

func decimalPtr(value string) *decimal.Decimal {
	amount := decimal.RequireFromString(value)
	return &amount
}

func stringPtr(value string) *string {
	return &value
}

func TestNormalizeISBN(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		expectOK bool
	}{
		{name: "normalizes isbn13 separators", input: " 978-7 111128069 ", expected: "9787111128069", expectOK: true},
		{name: "accepts isbn10 digits", input: "0306406152", expected: "0306406152", expectOK: true},
		{name: "accepts isbn10 check x", input: "030640615x", expected: "030640615X", expectOK: true},
		{name: "rejects bad length", input: "978711112806", expectOK: false},
		{name: "rejects letters inside isbn13", input: "978711112806X", expectOK: false},
		{name: "rejects x outside isbn10 check digit", input: "03064X6152", expectOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, ok := normalizeISBN(tt.input)
			if ok != tt.expectOK {
				t.Fatalf("ok = %v, want %v", ok, tt.expectOK)
			}
			if actual != tt.expected {
				t.Fatalf("normalized = %q, want %q", actual, tt.expected)
			}
		})
	}
}

func TestDecideRecycle(t *testing.T) {
	book := models.Book{Isbn: "9787111128069", Title: "示例可回收图书", Source: "mock"}

	t.Run("rejects unknown scanned placeholder book", func(t *testing.T) {
		decision, reason, suggested := decideRecycle(models.Book{Isbn: book.Isbn, Source: "scan"}, false, false, models.PriceSnapshot{})
		if decision != decisionReject {
			t.Fatalf("decision = %s, want %s", decision, decisionReject)
		}
		if reason != "未找到书籍基础信息" {
			t.Fatalf("reason = %q", reason)
		}
		if suggested != nil {
			t.Fatalf("suggested price = %v, want nil", suggested)
		}
	})

	t.Run("accepts existing book without valid price", func(t *testing.T) {
		decision, _, suggested := decideRecycle(book, true, false, models.PriceSnapshot{})
		if decision != decisionAccept {
			t.Fatalf("decision = %s, want %s", decision, decisionAccept)
		}
		if suggested != nil {
			t.Fatalf("suggested price = %v, want nil", suggested)
		}
	})

	t.Run("accepts existing book with low sample count", func(t *testing.T) {
		price := models.PriceSnapshot{AvgPrice: decimalPtr("30.00"), SampleCount: 1, Confidence: "LOW"}
		decision, _, suggested := decideRecycle(book, true, true, price)
		if decision != decisionAccept {
			t.Fatalf("decision = %s, want %s", decision, decisionAccept)
		}
		if suggested == nil || !suggested.Equal(decimal.RequireFromString("9.00")) {
			t.Fatalf("suggested price = %v, want 9.00", suggested)
		}
	})

	t.Run("accepts existing book with low average price", func(t *testing.T) {
		price := models.PriceSnapshot{AvgPrice: decimalPtr("8.00"), SampleCount: 6, Confidence: "MEDIUM"}
		decision, _, suggested := decideRecycle(book, true, true, price)
		if decision != decisionAccept {
			t.Fatalf("decision = %s, want %s", decision, decisionAccept)
		}
		if suggested == nil || !suggested.Equal(decimal.RequireFromString("2.40")) {
			t.Fatalf("suggested price = %v, want 2.40", suggested)
		}
	})

	t.Run("accepts sufficient average price and samples", func(t *testing.T) {
		price := models.PriceSnapshot{AvgPrice: decimalPtr("24.50"), SampleCount: 8, Confidence: "HIGH"}
		decision, _, suggested := decideRecycle(book, true, true, price)
		if decision != decisionAccept {
			t.Fatalf("decision = %s, want %s", decision, decisionAccept)
		}
		if suggested == nil {
			t.Fatal("suggested price = nil, want decimal")
		}
		if !suggested.Equal(decimal.RequireFromString("7.35")) {
			t.Fatalf("suggested price = %s, want 7.35", suggested.String())
		}
	})
}

func TestDecideRecycleUsesEnabledRuleFromDatabase(t *testing.T) {
	db := setupControllerTestDB(t)
	now := time.Now()
	rule := models.RecycleRule{
		ID:                1,
		Version:           "v-test",
		Name:              "测试规则",
		MinAcceptAvgPrice: decimal.RequireFromString("20.00"),
		RecycleRate:       decimal.RequireFromString("0.25"),
		MinSampleCount:    5,
		LowConfidenceMode: "NEED_REVIEW",
		Enabled:           true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := db.Create(&rule).Error; err != nil {
		t.Fatalf("create rule: %v", err)
	}

	book := models.Book{Isbn: "9787111128069", Title: "示例可回收图书", Source: "mock"}
	price := models.PriceSnapshot{AvgPrice: decimalPtr("40.00"), SampleCount: 6, Confidence: "HIGH"}
	decision, _, suggested := decideRecycle(book, true, true, price)

	if decision != decisionAccept {
		t.Fatalf("decision = %s, want %s", decision, decisionAccept)
	}
	if suggested == nil || !suggested.Equal(decimal.RequireFromString("10.00")) {
		t.Fatalf("suggested price = %v, want 10.00", suggested)
	}
}

func TestRecordNotFoundHelpersReturnEmptyResults(t *testing.T) {
	setupControllerTestDB(t)
	controller := &BookController{}
	now := time.Now()

	if _, found, err := findBook("9787111128069"); err != nil || found {
		t.Fatalf("findBook found=%v err=%v, want found=false err=nil", found, err)
	}

	if _, hasPrice, err := latestValidPrice("9787111128069", now); err != nil || hasPrice {
		t.Fatalf("latestValidPrice hasPrice=%v err=%v, want hasPrice=false err=nil", hasPrice, err)
	}

	if _, hasScanLog, err := latestScanLog("9787111128069"); err != nil || hasScanLog {
		t.Fatalf("latestScanLog hasScanLog=%v err=%v, want hasScanLog=false err=nil", hasScanLog, err)
	}

	lastScannedAt, scannedRecently, duplicateInBatch, err := duplicateInfo("9787111128069", "batch_001", now)
	if err != nil {
		t.Fatalf("duplicateInfo err=%v, want nil", err)
	}
	if lastScannedAt != nil || scannedRecently || duplicateInBatch {
		t.Fatalf("duplicateInfo = (%v, %v, %v), want nil,false,false", lastScannedAt, scannedRecently, duplicateInBatch)
	}

	missingPriceID := uint64(999)
	if _, hasPrice, err := priceByID(&missingPriceID); err != nil || hasPrice {
		t.Fatalf("priceByID hasPrice=%v err=%v, want hasPrice=false err=nil", hasPrice, err)
	}

	if _, found, err := controller.findIdempotentCheckResponse("missing_request"); err != nil || found {
		t.Fatalf("findIdempotentCheckResponse found=%v err=%v, want found=false err=nil", found, err)
	}
}

func TestFindIdempotentCheckResponseReusesExistingScanLog(t *testing.T) {
	db := setupControllerTestDB(t)
	controller := &BookController{}
	now := time.Now()
	clientRequestID := "request_001"

	book := models.Book{
		Isbn:      "9787111128069",
		Title:     "示例可回收图书",
		Source:    "mock",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&book).Error; err != nil {
		t.Fatalf("create book: %v", err)
	}

	scanLog := models.ScanLog{
		ID:                    1,
		Isbn:                  book.Isbn,
		NormalizedIsbn:        book.Isbn,
		Decision:              decisionAccept,
		Reason:                "二手市场均价满足回收规则",
		SuggestedRecyclePrice: decimalPtr("7.35"),
		Confidence:            "HIGH",
		ClientRequestID:       &clientRequestID,
		ScannedAt:             now,
		CreatedAt:             now,
	}
	if err := db.Create(&scanLog).Error; err != nil {
		t.Fatalf("create scan log: %v", err)
	}

	resp, found, err := controller.findIdempotentCheckResponse(clientRequestID)
	if err != nil {
		t.Fatalf("findIdempotentCheckResponse err=%v, want nil", err)
	}
	if !found {
		t.Fatal("findIdempotentCheckResponse found=false, want true")
	}
	if resp.ScanLogID != scanLog.ID {
		t.Fatalf("scan_log_id = %d, want %d", resp.ScanLogID, scanLog.ID)
	}
	if resp.Book.Title != book.Title {
		t.Fatalf("book title = %q, want %q", resp.Book.Title, book.Title)
	}
	if resp.Decision != decisionAccept {
		t.Fatalf("decision = %s, want %s", resp.Decision, decisionAccept)
	}
}
