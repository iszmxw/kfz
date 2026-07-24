package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"goapi/app/models"
	"goapi/pkg/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const KongfzCategorySource = "kongfz_category"

type KongfzCategorySyncInput struct {
	CatID     int
	Source    string
	CreatedBy uint64
	Items     []KongfzCategoryItemInput
}

type KongfzCategoryItemInput struct {
	RowNumber        int
	Page             int
	Isbn             string
	Title            string
	BookName         string
	Author           string
	Publisher        string
	PublishYear      string
	PublishDate      string
	Binding          string
	ListPrice        string
	CoverURL         string
	RawURL           string
	KongfzID         uint64
	Mid              uint64
	OldBookMinPrice  *decimal.Decimal
	OldBookOnSaleNum int
	BookShowInfo     []string
	RawPayload       string
}

type KongfzCategorySyncSummary struct {
	Total            int `json:"total"`
	Success          int `json:"success"`
	Failed           int `json:"failed"`
	CreatedBooks     int `json:"created_books"`
	UpdatedBooks     int `json:"updated_books"`
	CreatedSnapshots int `json:"created_snapshots"`
}

type KongfzCategorySyncRow struct {
	RowNumber            int    `json:"row_number"`
	Isbn                 string `json:"isbn"`
	Title                string `json:"title"`
	Valid                bool   `json:"valid"`
	ErrorMessage         string `json:"error_message"`
	BookCreated          bool   `json:"book_created"`
	BookUpdated          bool   `json:"book_updated"`
	PriceSnapshotCreated bool   `json:"price_snapshot_created"`
	PriceSnapshotID      uint64 `json:"price_snapshot_id"`
	ImportTaskRowID      uint64 `json:"import_task_row_id"`
}

type KongfzCategorySyncResult struct {
	Task    models.ImportTask         `json:"task"`
	Summary KongfzCategorySyncSummary `json:"summary"`
	Rows    []KongfzCategorySyncRow   `json:"rows"`
}

type normalizedKongfzItem struct {
	Input       KongfzCategoryItemInput
	RowNumber   int
	Isbn        string
	Title       string
	Author      string
	Publisher   string
	PublishYear string
	CoverURL    string
	RawPayload  string
	Valid       bool
	Error       string
}

func SyncKongfzCategory(input KongfzCategorySyncInput) (KongfzCategorySyncResult, error) {
	return SyncKongfzCategoryWithDB(mysql.DB, input)
}

func SyncKongfzCategoryWithDB(db *gorm.DB, input KongfzCategorySyncInput) (KongfzCategorySyncResult, error) {
	if input.CatID <= 0 {
		return KongfzCategorySyncResult{}, errors.New("cat_id无效")
	}
	if len(input.Items) == 0 {
		return KongfzCategorySyncResult{}, errors.New("同步数据不能为空")
	}
	if input.Source == "" {
		input.Source = KongfzCategorySource
	}

	now := time.Now()
	plans := normalizeKongfzItems(input)
	result := KongfzCategorySyncResult{
		Summary: KongfzCategorySyncSummary{Total: len(plans)},
		Rows:    make([]KongfzCategorySyncRow, 0, len(plans)),
	}
	for _, plan := range plans {
		if plan.Valid {
			result.Summary.Success++
		} else {
			result.Summary.Failed++
		}
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		task := models.ImportTask{
			FileName:    fmt.Sprintf("kongfz_category_cat_%d_%s.json", input.CatID, now.Format("20060102150405")),
			FileType:    "json",
			Status:      "COMPLETED",
			TotalRows:   result.Summary.Total,
			SuccessRows: result.Summary.Success,
			FailedRows:  result.Summary.Failed,
			CreatedBy:   input.CreatedBy,
			CreatedAt:   now,
			CompletedAt: &now,
		}
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
		result.Task = task

		for _, plan := range plans {
			row := models.ImportTaskRow{
				TaskID:       task.ID,
				RowNumber:    plan.RowNumber,
				Isbn:         plan.Isbn,
				Title:        plan.Title,
				Valid:        plan.Valid,
				ErrorMessage: plan.Error,
				RawPayload:   plan.RawPayload,
				CreatedAt:    now,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}

			out := KongfzCategorySyncRow{
				RowNumber:       plan.RowNumber,
				Isbn:            plan.Isbn,
				Title:           plan.Title,
				Valid:           plan.Valid,
				ErrorMessage:    plan.Error,
				ImportTaskRowID: row.ID,
			}
			if plan.Valid {
				bookCreated, bookUpdated, err := upsertKongfzBook(tx, plan, input.Source, now)
				if err != nil {
					return err
				}
				out.BookCreated = bookCreated
				out.BookUpdated = bookUpdated
				if bookCreated {
					result.Summary.CreatedBooks++
				}
				if bookUpdated {
					result.Summary.UpdatedBooks++
				}

				priceSnapshotID, created, err := createKongfzPriceSnapshot(tx, task.ID, row.ID, plan, input.Source, now)
				if err != nil {
					return err
				}
				out.PriceSnapshotID = priceSnapshotID
				out.PriceSnapshotCreated = created
				if created {
					result.Summary.CreatedSnapshots++
				}
			}
			result.Rows = append(result.Rows, out)
		}
		return nil
	})
	if err != nil {
		return KongfzCategorySyncResult{}, err
	}
	return result, nil
}

func normalizeKongfzItems(input KongfzCategorySyncInput) []normalizedKongfzItem {
	seen := map[string]struct{}{}
	items := make([]normalizedKongfzItem, 0, len(input.Items))
	for index, item := range input.Items {
		rowNumber := item.RowNumber
		if rowNumber <= 0 {
			rowNumber = index + 1
		}

		title := strings.TrimSpace(firstNonEmpty(item.Title, item.BookName))
		bookShowInfo := item.BookShowInfo
		author := strings.TrimSpace(item.Author)
		if author == "" && len(bookShowInfo) > 0 {
			author = strings.TrimSpace(bookShowInfo[0])
		}
		publisher := strings.TrimSpace(item.Publisher)
		if publisher == "" && len(bookShowInfo) > 1 {
			publisher = strings.TrimSpace(bookShowInfo[1])
		}
		publishYear := strings.TrimSpace(item.PublishYear)
		if publishYear == "" && item.PublishDate != "" {
			publishYear = extractPublishYear(item.PublishDate)
		}
		if publishYear == "" && len(bookShowInfo) > 2 {
			publishYear = extractPublishYear(bookShowInfo[2])
		}

		isbn := NormalizeISBN(item.Isbn)
		rawPayload := normalizedRawPayload(item.RawPayload, item)
		var rowErrors []string
		if isbn == "" {
			rowErrors = append(rowErrors, "ISBN格式无效")
		}
		if title == "" {
			rowErrors = append(rowErrors, "书名不能为空")
		}
		if isbn != "" && title != "" {
			if _, exists := seen[isbn]; exists {
				rowErrors = append(rowErrors, "本次同步中ISBN重复")
			} else {
				seen[isbn] = struct{}{}
			}
		}

		items = append(items, normalizedKongfzItem{
			Input:       item,
			RowNumber:   rowNumber,
			Isbn:        isbn,
			Title:       title,
			Author:      author,
			Publisher:   publisher,
			PublishYear: publishYear,
			CoverURL:    strings.TrimSpace(item.CoverURL),
			RawPayload:  rawPayload,
			Valid:       len(rowErrors) == 0,
			Error:       strings.Join(rowErrors, "；"),
		})
	}
	return items
}

func upsertKongfzBook(tx *gorm.DB, item normalizedKongfzItem, source string, now time.Time) (bool, bool, error) {
	var existing models.Book
	find := tx.Where(models.BookColumns.Isbn+" = ?", item.Isbn).Take(&existing)
	if errors.Is(find.Error, gorm.ErrRecordNotFound) || find.RowsAffected == 0 {
		book := models.Book{
			Isbn:        item.Isbn,
			Title:       item.Title,
			Author:      item.Author,
			Publisher:   item.Publisher,
			PublishYear: item.PublishYear,
			CoverURL:    item.CoverURL,
			Source:      source,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: models.BookColumns.Isbn}},
			DoNothing: true,
		}).Create(&book).Error
		if err != nil {
			return false, false, err
		}
		if book.ID != 0 {
			return true, false, nil
		}
		find = tx.Where(models.BookColumns.Isbn+" = ?", item.Isbn).Take(&existing)
	}
	if find.Error != nil {
		return false, false, find.Error
	}

	updates := map[string]interface{}{}
	canFill := func(value string) bool {
		return strings.TrimSpace(value) == "" || existing.Source == "scan"
	}
	if item.Title != "" && canFill(existing.Title) && existing.Title != item.Title {
		updates[models.BookColumns.Title] = item.Title
	}
	if item.Author != "" && canFill(existing.Author) && existing.Author != item.Author {
		updates[models.BookColumns.Author] = item.Author
	}
	if item.Publisher != "" && canFill(existing.Publisher) && existing.Publisher != item.Publisher {
		updates[models.BookColumns.Publisher] = item.Publisher
	}
	if item.PublishYear != "" && canFill(existing.PublishYear) && existing.PublishYear != item.PublishYear {
		updates[models.BookColumns.PublishYear] = item.PublishYear
	}
	if item.CoverURL != "" && canFill(existing.CoverURL) && existing.CoverURL != item.CoverURL {
		updates[models.BookColumns.CoverURL] = item.CoverURL
	}
	if (existing.Source == "" || existing.Source == "scan") && existing.Source != source {
		updates[models.BookColumns.Source] = source
	}
	if len(updates) == 0 {
		return false, false, nil
	}
	updates[models.BookColumns.UpdatedAt] = now
	return false, true, tx.Model(&existing).Updates(updates).Error
}

func createKongfzPriceSnapshot(tx *gorm.DB, taskID, rowID uint64, item normalizedKongfzItem, source string, now time.Time) (uint64, bool, error) {
	if item.Input.OldBookMinPrice == nil || !item.Input.OldBookMinPrice.GreaterThan(decimal.Zero) {
		return 0, false, nil
	}
	sampleCount := item.Input.OldBookOnSaleNum
	if sampleCount < 0 {
		sampleCount = 0
	}
	price := models.PriceSnapshot{
		Isbn:          item.Isbn,
		Source:        source,
		MinPrice:      item.Input.OldBookMinPrice,
		SampleCount:   sampleCount,
		Confidence:    "LOW",
		RawURL:        strings.TrimSpace(item.Input.RawURL),
		RawPayloadRef: fmt.Sprintf("import_task:%d:row:%d", taskID, rowID),
		CollectedAt:   now,
		CreatedAt:     now,
	}
	if err := tx.Create(&price).Error; err != nil {
		return 0, false, err
	}
	return price.ID, true, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func extractPublishYear(value string) string {
	value = strings.TrimSpace(value)
	if len(value) < 4 {
		return value
	}
	year := value[:4]
	for _, ch := range year {
		if ch < '0' || ch > '9' {
			return value
		}
	}
	return year
}

func normalizedRawPayload(raw string, fallback KongfzCategoryItemInput) string {
	raw = strings.TrimSpace(raw)
	if raw != "" && json.Valid([]byte(raw)) {
		return raw
	}
	payload, err := json.Marshal(fallback)
	if err != nil {
		return "{}"
	}
	return string(payload)
}
