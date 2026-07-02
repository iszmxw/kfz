package v1

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"goapi/app/models"
	"goapi/app/requests"
	"goapi/app/response"
	"goapi/pkg/echo"
	"goapi/pkg/helpers"
	"goapi/pkg/mysql"
	"gorm.io/gorm"
)

const (
	decisionAccept     = "ACCEPT"
	decisionReject     = "REJECT"
	decisionNeedReview = "NEED_REVIEW"
	confidenceNone     = "NONE"
)

var (
	minAcceptAvgPrice = decimal.RequireFromString("10.00")
	recycleRate       = decimal.RequireFromString("0.30")
)

type BookController struct {
	BaseController
}

func (h *BookController) Check(c *gin.Context) {
	req := requests.BookCheck{
		Isbn:            c.Query("isbn"),
		BatchID:         c.Query("batch_id"),
		ClientRequestID: c.Query("client_request_id"),
	}
	normalizedISBN, ok := normalizeISBN(req.Isbn)
	if !ok {
		echo.Error(c, "Failed", "未识别到有效 ISBN")
		return
	}

	if req.ClientRequestID != "" {
		if resp, found, err := h.findIdempotentCheckResponse(req.ClientRequestID); err != nil {
			echo.Error(c, "Failed", err.Error())
			return
		} else if found {
			echo.Success(c, resp, "")
			return
		}
	}

	now := time.Now()
	book, bookExisted, err := h.ensureBook(normalizedISBN, now)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}

	price, hasPrice, err := latestValidPrice(normalizedISBN, now)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}

	lastScan, scannedRecently, duplicateInBatch, err := duplicateInfo(normalizedISBN, req.BatchID, now)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}

	decision, reason, suggestedPrice := decideRecycle(book, bookExisted, hasPrice, price)
	scanLog := buildScanLog(req, normalizedISBN, decision, reason, suggestedPrice, price, hasPrice, scannedRecently, duplicateInBatch, now)
	if err := mysql.DB.Create(&scanLog).Error; err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}

	resp := buildBookCheckResponse(book, scanLog, price, hasPrice, lastScan, duplicateInBatch)
	echo.Success(c, resp, "")
}

func (h *BookController) Detail(c *gin.Context) {
	req := requests.BookDetail{Isbn: c.Query("isbn")}
	normalizedISBN, ok := normalizeISBN(req.Isbn)
	if !ok {
		echo.Error(c, "Failed", "未识别到有效 ISBN")
		return
	}

	book, found, err := findBook(normalizedISBN)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	if !found {
		echo.Error(c, "Failed", "未找到图书信息")
		return
	}

	now := time.Now()
	price, hasPrice, err := latestValidPrice(normalizedISBN, now)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}

	scanLog, hasScanLog, err := latestScanLog(normalizedISBN)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}

	decision, reason, suggestedPrice := decideRecycle(book, true, hasPrice, price)
	updatedAt := now
	if hasScanLog {
		decision = scanLog.Decision
		reason = scanLog.Reason
		suggestedPrice = scanLog.SuggestedRecyclePrice
		updatedAt = scanLog.ScannedAt
	}

	resp := response.BookDetailResponse{
		Book:              buildBookDTO(book),
		LatestMarketPrice: buildMarketPriceDTO(price, hasPrice),
		LatestDecision: response.LatestDecisionDTO{
			Decision:              decision,
			Reason:                reason,
			SuggestedRecyclePrice: decimalToFloatPtr(suggestedPrice),
			UpdatedAt:             formatTime(updatedAt),
		},
	}
	echo.Success(c, resp, "")
}

func (h *BookController) ensureBook(normalizedISBN string, now time.Time) (models.Book, bool, error) {
	book, found, err := findBook(normalizedISBN)
	if err != nil {
		return book, false, err
	}
	if found {
		return book, true, nil
	}

	book = models.Book{
		Isbn:      normalizedISBN,
		Title:     "未知图书",
		Source:    "scan",
		CreatedAt: now,
		UpdatedAt: now,
	}
	return book, false, mysql.DB.Create(&book).Error
}

func (h *BookController) findIdempotentCheckResponse(clientRequestID string) (response.BookCheckResponse, bool, error) {
	var scanLog models.ScanLog
	tx := mysql.DB.Where(models.ScanLogColumns.ClientRequestID+" = ?", clientRequestID).First(&scanLog)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return response.BookCheckResponse{}, false, nil
	}
	if tx.Error != nil {
		return response.BookCheckResponse{}, false, tx.Error
	}
	if tx.RowsAffected == 0 {
		return response.BookCheckResponse{}, false, nil
	}

	book, _, err := findBook(scanLog.NormalizedIsbn)
	if err != nil {
		return response.BookCheckResponse{}, false, err
	}
	price, hasPrice, err := priceByID(scanLog.PriceSnapshotID)
	if err != nil {
		return response.BookCheckResponse{}, false, err
	}
	resp := buildBookCheckResponse(book, scanLog, price, hasPrice, nil, scanLog.DuplicateInBatch)
	return resp, true, nil
}

func findBook(normalizedISBN string) (models.Book, bool, error) {
	var book models.Book
	tx := mysql.DB.Where(models.BookColumns.Isbn+" = ?", normalizedISBN).First(&book)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return book, false, nil
	}
	if tx.Error != nil {
		return book, false, tx.Error
	}
	return book, tx.RowsAffected > 0, nil
}

func latestValidPrice(normalizedISBN string, now time.Time) (models.PriceSnapshot, bool, error) {
	var price models.PriceSnapshot
	tx := mysql.DB.
		Where(models.PriceSnapshotColumns.Isbn+" = ?", normalizedISBN).
		Where("("+models.PriceSnapshotColumns.ExpiresAt+" IS NULL OR "+models.PriceSnapshotColumns.ExpiresAt+" > ?)", now).
		Order(models.PriceSnapshotColumns.CollectedAt + " DESC").
		First(&price)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return price, false, nil
	}
	if tx.Error != nil {
		return price, false, tx.Error
	}
	return price, tx.RowsAffected > 0, nil
}

func priceByID(priceSnapshotID *string) (models.PriceSnapshot, bool, error) {
	if priceSnapshotID == nil || *priceSnapshotID == "" {
		return models.PriceSnapshot{}, false, nil
	}

	var price models.PriceSnapshot
	tx := mysql.DB.Where(models.PriceSnapshotColumns.ID+" = ?", *priceSnapshotID).First(&price)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return price, false, nil
	}
	if tx.Error != nil {
		return price, false, tx.Error
	}
	return price, tx.RowsAffected > 0, nil
}

func latestScanLog(normalizedISBN string) (models.ScanLog, bool, error) {
	var scanLog models.ScanLog
	tx := mysql.DB.
		Where(models.ScanLogColumns.NormalizedIsbn+" = ?", normalizedISBN).
		Order(models.ScanLogColumns.ScannedAt + " DESC").
		First(&scanLog)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return scanLog, false, nil
	}
	if tx.Error != nil {
		return scanLog, false, tx.Error
	}
	return scanLog, tx.RowsAffected > 0, nil
}

func duplicateInfo(normalizedISBN, batchID string, now time.Time) (*time.Time, bool, bool, error) {
	var lastLog models.ScanLog
	tx := mysql.DB.
		Where(models.ScanLogColumns.NormalizedIsbn+" = ?", normalizedISBN).
		Order(models.ScanLogColumns.ScannedAt + " DESC").
		First(&lastLog)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		tx.Error = nil
	}
	if tx.Error != nil {
		return nil, false, false, tx.Error
	}

	var lastScannedAt *time.Time
	scannedRecently := false
	if tx.RowsAffected > 0 {
		lastScannedAt = &lastLog.ScannedAt
		scannedRecently = now.Sub(lastLog.ScannedAt) <= 10*time.Minute
	}

	duplicateInBatch := false
	if batchID != "" {
		var count int64
		countTx := mysql.DB.Model(&models.ScanLog{}).
			Where(models.ScanLogColumns.NormalizedIsbn+" = ?", normalizedISBN).
			Where(models.ScanLogColumns.BatchID+" = ?", batchID).
			Count(&count)
		if countTx.Error != nil {
			return nil, false, false, countTx.Error
		}
		duplicateInBatch = count > 0
	}

	return lastScannedAt, scannedRecently, duplicateInBatch, nil
}

func buildScanLog(req requests.BookCheck, normalizedISBN, decision, reason string, suggestedPrice *decimal.Decimal, price models.PriceSnapshot, hasPrice bool, scannedRecently, duplicateInBatch bool, now time.Time) models.ScanLog {
	var batchID *string
	if req.BatchID != "" {
		batchID = &req.BatchID
	}
	var clientRequestID *string
	if req.ClientRequestID != "" {
		clientRequestID = &req.ClientRequestID
	}
	var marketSampleCount *int
	var priceSnapshotID *string
	var minPrice, avgPrice, maxPrice *decimal.Decimal
	confidence := confidenceNone
	if hasPrice {
		minPrice = price.MinPrice
		avgPrice = price.AvgPrice
		maxPrice = price.MaxPrice
		marketSampleCount = &price.SampleCount
		confidence = price.Confidence
		priceSnapshotID = &price.ID
	}

	return models.ScanLog{
		ID:                    helpers.GetUUID(),
		Isbn:                  req.Isbn,
		NormalizedIsbn:        normalizedISBN,
		BatchID:               batchID,
		Decision:              decision,
		Reason:                reason,
		MarketMinPrice:        minPrice,
		MarketAvgPrice:        avgPrice,
		MarketMaxPrice:        maxPrice,
		MarketSampleCount:     marketSampleCount,
		SuggestedRecyclePrice: suggestedPrice,
		Confidence:            confidence,
		PriceSnapshotID:       priceSnapshotID,
		DuplicateRecently:     scannedRecently,
		DuplicateInBatch:      duplicateInBatch,
		ClientRequestID:       clientRequestID,
		ScannedAt:             now,
		CreatedAt:             now,
	}
}

func decideRecycle(book models.Book, bookExisted bool, hasPrice bool, price models.PriceSnapshot) (string, string, *decimal.Decimal) {
	if !bookExisted || book.Source == "scan" {
		return decisionReject, "未找到书籍基础信息", nil
	}
	if !hasPrice || price.AvgPrice == nil || price.SampleCount <= 0 {
		return decisionNeedReview, "暂无有效价格数据，需要人工确认", nil
	}
	if price.SampleCount < 3 || price.Confidence == "LOW" || price.Confidence == confidenceNone {
		return decisionNeedReview, "价格样本不足或可信度偏低，需要人工确认", nil
	}
	if price.AvgPrice.LessThan(minAcceptAvgPrice) {
		return decisionReject, "二手市场均价低于回收规则", nil
	}

	suggested := price.AvgPrice.Mul(recycleRate).Round(2)
	return decisionAccept, "二手市场均价满足回收规则", &suggested
}

func buildBookCheckResponse(book models.Book, scanLog models.ScanLog, price models.PriceSnapshot, hasPrice bool, lastScannedAt *time.Time, duplicateInBatch bool) response.BookCheckResponse {
	resp := response.BookCheckResponse{
		ScanLogID:             scanLog.ID,
		Book:                  buildBookDTO(book),
		Decision:              scanLog.Decision,
		Reason:                scanLog.Reason,
		SuggestedRecyclePrice: decimalToFloatPtr(scanLog.SuggestedRecyclePrice),
		MarketPrice:           buildMarketPriceDTO(price, hasPrice),
		Duplicate: response.DuplicateInfoDTO{
			ScannedRecently:         scanLog.DuplicateRecently,
			DuplicateInCurrentBatch: duplicateInBatch,
			LastScannedAt:           formatTimePtr(lastScannedAt),
		},
		ScannedAt: formatTime(scanLog.ScannedAt),
	}
	if duplicateInBatch {
		resp.BatchActionRequired = "CONFIRM_DUPLICATE"
	}
	return resp
}

func buildBookDTO(book models.Book) response.BookDTO {
	return response.BookDTO{
		Isbn:           book.Isbn,
		NormalizedIsbn: book.Isbn,
		Title:          book.Title,
		Author:         book.Author,
		Publisher:      book.Publisher,
		PublishYear:    book.PublishYear,
		CoverURL:       book.CoverURL,
	}
}

func buildMarketPriceDTO(price models.PriceSnapshot, hasPrice bool) *response.MarketPriceDTO {
	if !hasPrice {
		return nil
	}
	collectedAt := formatTime(price.CollectedAt)
	return &response.MarketPriceDTO{
		Source:      price.Source,
		Min:         decimalToFloatPtr(price.MinPrice),
		Avg:         decimalToFloatPtr(price.AvgPrice),
		Max:         decimalToFloatPtr(price.MaxPrice),
		SampleCount: price.SampleCount,
		Confidence:  price.Confidence,
		CollectedAt: &collectedAt,
	}
}

func normalizeISBN(isbn string) (string, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(isbn))
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, " ", "")
	if len(normalized) == 10 {
		for i := 0; i < 10; i++ {
			ch := normalized[i]
			if ch >= '0' && ch <= '9' {
				continue
			}
			if i == 9 && ch == 'X' {
				continue
			}
			return "", false
		}
		return normalized, true
	}
	if len(normalized) == 13 {
		for i := 0; i < 13; i++ {
			ch := normalized[i]
			if ch < '0' || ch > '9' {
				return "", false
			}
		}
		return normalized, true
	}
	return "", false
}

func decimalValue(value *decimal.Decimal) (float64, bool) {
	if value == nil {
		return 0, false
	}
	result, _ := value.Float64()
	return result, true
}

func decimalToFloatPtr(decimalValuePtr *decimal.Decimal) *float64 {
	value, ok := decimalValue(decimalValuePtr)
	if !ok {
		return nil
	}
	result := helpers.Decimal(value, 2)
	return &result
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := formatTime(*t)
	return &formatted
}
