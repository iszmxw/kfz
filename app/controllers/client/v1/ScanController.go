package v1

import (
	"strings"

	"github.com/gin-gonic/gin"
	"goapi/app/models"
	"goapi/app/requests"
	"goapi/app/response"
	"goapi/pkg/echo"
	"goapi/pkg/mysql"
	"gorm.io/gorm"
)

type ScanController struct {
	BaseController
}

func (h *ScanController) History(c *gin.Context) {
	req := requests.ScanHistory{
		Page:     parsePositiveInt(c.Query("page"), 1),
		PageSize: parsePositiveInt(c.Query("page_size"), 20),
		Decision: strings.ToUpper(strings.TrimSpace(c.Query("decision"))),
		Isbn:     c.Query("isbn"),
		BatchID:  c.Query("batch_id"),
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.Decision != "" && !validDecision(req.Decision) {
		echo.Error(c, "Failed", "检查参数decision")
		return
	}

	normalizedISBN := ""
	if req.Isbn != "" {
		var ok bool
		normalizedISBN, ok = normalizeISBN(req.Isbn)
		if !ok {
			echo.Error(c, "Failed", "未识别到有效 ISBN")
			return
		}
	}

	var total int64
	if err := scanHistoryQuery(req, normalizedISBN).Count(&total).Error; err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}

	var logs []models.ScanLog
	offset := (req.Page - 1) * req.PageSize
	if err := scanHistoryQuery(req, normalizedISBN).
		Order(models.ScanLogColumns.ScannedAt + " DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&logs).Error; err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}

	books, err := findBooksByScanLogs(logs)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}

	items := make([]response.ScanHistoryItem, 0, len(logs))
	for _, log := range logs {
		title := ""
		if book, ok := books[log.NormalizedIsbn]; ok {
			title = book.Title
		}
		items = append(items, response.ScanHistoryItem{
			ScanLogID:             log.ID,
			Isbn:                  log.NormalizedIsbn,
			Title:                 title,
			Decision:              log.Decision,
			Reason:                log.Reason,
			SuggestedRecyclePrice: decimalToFloatPtr(log.SuggestedRecyclePrice),
			Confidence:            log.Confidence,
			ScannedAt:             formatTime(log.ScannedAt),
		})
	}

	echo.Success(c, response.ScanHistoryResponse{
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
		Items:    items,
	}, "")
}

func scanHistoryQuery(req requests.ScanHistory, normalizedISBN string) *gorm.DB {
	query := mysql.DB.Model(&models.ScanLog{})
	if req.Decision != "" {
		query = query.Where(models.ScanLogColumns.Decision+" = ?", req.Decision)
	}
	if normalizedISBN != "" {
		query = query.Where(models.ScanLogColumns.NormalizedIsbn+" = ?", normalizedISBN)
	}
	if req.BatchID != "" {
		query = query.Where(models.ScanLogColumns.BatchID+" = ?", req.BatchID)
	}
	return query
}

func findBooksByScanLogs(logs []models.ScanLog) (map[string]models.Book, error) {
	result := make(map[string]models.Book)
	if len(logs) == 0 {
		return result, nil
	}

	isbnSet := make(map[string]struct{})
	isbns := make([]string, 0, len(logs))
	for _, log := range logs {
		if _, ok := isbnSet[log.NormalizedIsbn]; ok {
			continue
		}
		isbnSet[log.NormalizedIsbn] = struct{}{}
		isbns = append(isbns, log.NormalizedIsbn)
	}

	var books []models.Book
	if err := mysql.DB.Where(models.BookColumns.Isbn+" IN ?", isbns).Find(&books).Error; err != nil {
		return result, err
	}
	for _, book := range books {
		result[book.Isbn] = book
	}
	return result, nil
}

func validDecision(decision string) bool {
	return decision == decisionAccept || decision == decisionReject || decision == decisionNeedReview
}

func parsePositiveInt(value string, fallback int) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	var result int
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return fallback
		}
		result = result*10 + int(ch-'0')
	}
	if result <= 0 {
		return fallback
	}
	return result
}
