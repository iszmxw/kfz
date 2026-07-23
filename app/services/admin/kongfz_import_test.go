package admin

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"goapi/app/models"
)

func TestSyncKongfzCategoryCreatesBookAndConservativePrice(t *testing.T) {
	db := setupAdminServiceTestDB(t)
	minPrice := decimal.RequireFromString("0.01")

	result, err := SyncKongfzCategory(KongfzCategorySyncInput{
		CatID:     43,
		CreatedBy: 1,
		Items: []KongfzCategoryItemInput{
			{
				Isbn:             "9787506365437",
				BookName:         "活着",
				CoverURL:         "https://example.com/cover.jpg",
				RawURL:           "https://search.kongfz.com/booklib/category?catId=43&page=1",
				OldBookMinPrice:  &minPrice,
				OldBookOnSaleNum: 4708,
				BookShowInfo:     []string{"余华  著", "作家出版社", "2012-08", "平装", "20.00"},
				RawPayload:       `{"isbn":"9787506365437","bookName":"活着"}`,
			},
		},
	})
	if err != nil {
		t.Fatalf("SyncKongfzCategory err=%v", err)
	}
	if result.Summary.Total != 1 || result.Summary.Success != 1 || result.Summary.Failed != 0 {
		t.Fatalf("summary counts = %+v", result.Summary)
	}
	if result.Summary.CreatedBooks != 1 || result.Summary.CreatedSnapshots != 1 {
		t.Fatalf("created summary = %+v", result.Summary)
	}

	var book models.Book
	if err := db.First(&book, "isbn = ?", "9787506365437").Error; err != nil {
		t.Fatalf("load book: %v", err)
	}
	if book.Title != "活着" || book.Author != "余华  著" || book.Publisher != "作家出版社" || book.PublishYear != "2012" {
		t.Fatalf("book mapped incorrectly: %+v", book)
	}
	if book.Source != KongfzCategorySource {
		t.Fatalf("book source = %q, want %q", book.Source, KongfzCategorySource)
	}

	var price models.PriceSnapshot
	if err := db.First(&price, "isbn = ?", "9787506365437").Error; err != nil {
		t.Fatalf("load price snapshot: %v", err)
	}
	if price.MinPrice == nil || !price.MinPrice.Equal(minPrice) {
		t.Fatalf("min price = %v, want %s", price.MinPrice, minPrice)
	}
	if price.AvgPrice != nil || price.MaxPrice != nil {
		t.Fatalf("avg/max should be nil, got avg=%v max=%v", price.AvgPrice, price.MaxPrice)
	}
	if price.SampleCount != 4708 || price.Confidence != "LOW" {
		t.Fatalf("sample/confidence = %d/%s, want 4708/LOW", price.SampleCount, price.Confidence)
	}
	if !strings.Contains(price.RawPayloadRef, "import_task:") || !strings.Contains(price.RawPayloadRef, ":row:") {
		t.Fatalf("raw payload ref = %q", price.RawPayloadRef)
	}
}

func TestSyncKongfzCategoryFillsOnlyEmptyManualBookFields(t *testing.T) {
	db := setupAdminServiceTestDB(t)
	existing := models.Book{
		Isbn:      "9787020002207",
		Title:     "人工书名",
		Publisher: "人工出版社",
		Source:    "manual",
	}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("create existing book: %v", err)
	}

	_, err := SyncKongfzCategory(KongfzCategorySyncInput{
		CatID:     43,
		CreatedBy: 1,
		Items: []KongfzCategoryItemInput{
			{
				Isbn:        "9787020002207",
				BookName:    "红楼梦",
				Author:      "[清]曹雪芹、高鹗  著",
				Publisher:   "人民文学出版社",
				PublishDate: "2008-07",
				CoverURL:    "https://example.com/honglou.jpg",
				RawPayload:  `{"isbn":"9787020002207"}`,
			},
		},
	})
	if err != nil {
		t.Fatalf("SyncKongfzCategory err=%v", err)
	}

	var book models.Book
	if err := db.First(&book, "isbn = ?", "9787020002207").Error; err != nil {
		t.Fatalf("load book: %v", err)
	}
	if book.Title != "人工书名" {
		t.Fatalf("title overwritten to %q", book.Title)
	}
	if book.Publisher != "人工出版社" {
		t.Fatalf("publisher overwritten to %q", book.Publisher)
	}
	if book.Author != "[清]曹雪芹、高鹗  著" || book.PublishYear != "2008" || book.CoverURL == "" {
		t.Fatalf("empty fields were not filled: %+v", book)
	}
	if book.Source != "manual" {
		t.Fatalf("manual source overwritten to %q", book.Source)
	}
}

func TestSyncKongfzCategoryKeepsInvalidAndDuplicateRowsOutOfDomainTables(t *testing.T) {
	db := setupAdminServiceTestDB(t)
	minPrice := decimal.RequireFromString("0.20")

	result, err := SyncKongfzCategory(KongfzCategorySyncInput{
		CatID:     43,
		CreatedBy: 1,
		Items: []KongfzCategoryItemInput{
			{Isbn: "bad-isbn", BookName: "坏数据", RawPayload: `{"isbn":"bad-isbn"}`},
			{Isbn: "9787020002207", BookName: "红楼梦", OldBookMinPrice: &minPrice, OldBookOnSaleNum: 4823, RawPayload: `{"isbn":"9787020002207"}`},
			{Isbn: "9787020002207", BookName: "红楼梦重复", OldBookMinPrice: &minPrice, OldBookOnSaleNum: 4823, RawPayload: `{"isbn":"9787020002207","dup":true}`},
		},
	})
	if err != nil {
		t.Fatalf("SyncKongfzCategory err=%v", err)
	}
	if result.Summary.Total != 3 || result.Summary.Success != 1 || result.Summary.Failed != 2 {
		t.Fatalf("summary = %+v, want total=3 success=1 failed=2", result.Summary)
	}
	if result.Summary.CreatedBooks != 1 || result.Summary.CreatedSnapshots != 1 {
		t.Fatalf("created summary = %+v", result.Summary)
	}
	if result.Rows[0].Valid || !strings.Contains(result.Rows[0].ErrorMessage, "ISBN格式无效") {
		t.Fatalf("first row should be invalid ISBN: %+v", result.Rows[0])
	}
	if result.Rows[2].Valid || !strings.Contains(result.Rows[2].ErrorMessage, "ISBN重复") {
		t.Fatalf("duplicate row should be invalid: %+v", result.Rows[2])
	}

	var bookCount int64
	if err := db.Model(&models.Book{}).Count(&bookCount).Error; err != nil {
		t.Fatalf("count books: %v", err)
	}
	if bookCount != 1 {
		t.Fatalf("book count = %d, want 1", bookCount)
	}
	var snapshotCount int64
	if err := db.Model(&models.PriceSnapshot{}).Count(&snapshotCount).Error; err != nil {
		t.Fatalf("count snapshots: %v", err)
	}
	if snapshotCount != 1 {
		t.Fatalf("snapshot count = %d, want 1", snapshotCount)
	}
	var rowCount int64
	if err := db.Model(&models.ImportTaskRow{}).Count(&rowCount).Error; err != nil {
		t.Fatalf("count task rows: %v", err)
	}
	if rowCount != 3 {
		t.Fatalf("task row count = %d, want 3", rowCount)
	}
}
