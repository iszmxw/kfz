package v1

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"goapi/app/models"
	"goapi/app/response"
	adminSvc "goapi/app/services/admin"
	"goapi/pkg/echo"
	"goapi/pkg/mysql"
	"gorm.io/gorm"
)

type ManualReviewController struct{ BaseController }
type ScanLogController struct{ BaseController }
type BookController struct{ BaseController }
type PriceSnapshotController struct{ BaseController }
type ImportController struct{ BaseController }
type RecycleRuleController struct{ BaseController }

type kongfzCategorySyncRequest struct {
	CatID  int                             `json:"cat_id"`
	Source string                          `json:"source"`
	Items  []kongfzCategorySyncItemRequest `json:"items"`
}

type kongfzCategorySyncItemRequest struct {
	Isbn                string          `json:"isbn"`
	Title               string          `json:"title"`
	BookName            string          `json:"book_name"`
	Author              string          `json:"author"`
	Publisher           string          `json:"publisher"`
	PublishYear         string          `json:"publish_year"`
	PublishDate         string          `json:"publish_date"`
	Binding             string          `json:"binding"`
	ListPrice           string          `json:"list_price"`
	CoverURL            string          `json:"cover_url"`
	RawURL              string          `json:"raw_url"`
	Page                int             `json:"page"`
	Mid                 uint64          `json:"mid"`
	KongfzID            uint64          `json:"kongfz_id"`
	OldBookMinPrice     interface{}     `json:"old_book_min_price"`
	OldBookMinPriceText string          `json:"old_book_min_price_text"`
	OldBookOnSaleNum    int             `json:"old_book_on_sale_num"`
	BookShowInfo        []string        `json:"book_show_info"`
	RawPayload          json.RawMessage `json:"raw_payload"`
}

func (h *ManualReviewController) List(c *gin.Context) {
	page, pageSize := pageParams(c)
	query := mysql.DB.Model(&models.ScanLog{}).Where("decision = ?", "NEED_REVIEW")
	if c.Query("pending") != "false" {
		query = query.Where("NOT EXISTS (SELECT 1 FROM " + (&models.ManualDecision{}).TableName() + " md WHERE md.scan_log_id = " + (&models.ScanLog{}).TableName() + ".id)")
	}
	writePage(c, query.Order("scanned_at DESC"), page, pageSize, &[]models.ScanLog{})
}

func (h *ManualReviewController) Decide(c *gin.Context) {
	var req struct {
		ScanLogID          uint64  `json:"scan_log_id"`
		ManualDecision     string  `json:"manual_decision"`
		ActualRecyclePrice *string `json:"actual_recycle_price"`
		Note               string  `json:"note"`
	}
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	var price *decimal.Decimal
	if req.ActualRecyclePrice != nil && *req.ActualRecyclePrice != "" {
		amount, err := decimal.NewFromString(*req.ActualRecyclePrice)
		if err != nil {
			echo.Error(c, "Failed", "实际回收价格式错误")
			return
		}
		price = &amount
	}
	decision, err := adminSvc.CreateManualDecision(adminSvc.ManualDecisionInput{
		ScanLogID:          req.ScanLogID,
		AdminUserID:        currentSession(c).UserID,
		ManualDecision:     strings.ToUpper(req.ManualDecision),
		ActualRecyclePrice: price,
		Note:               req.Note,
	})
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "manual_review.decide", "manual_decision", "SUCCESS", req.ScanLogID)
	echo.Success(c, decision, "")
}

func (h *ScanLogController) List(c *gin.Context) {
	page, pageSize := pageParams(c)
	query := mysql.DB.Model(&models.ScanLog{})
	if decision := strings.ToUpper(strings.TrimSpace(c.Query("decision"))); decision != "" {
		query = query.Where("decision = ?", decision)
	}
	if isbn := strings.TrimSpace(c.Query("isbn")); isbn != "" {
		query = query.Where("normalized_isbn = ?", normalizeISBNForAdmin(isbn))
	}
	if batchID := strings.TrimSpace(c.Query("batch_id")); batchID != "" {
		query = query.Where("batch_id = ?", batchID)
	}
	writePage(c, query.Order("scanned_at DESC"), page, pageSize, &[]models.ScanLog{})
}

func (h *BookController) List(c *gin.Context) {
	page, pageSize := pageParams(c)
	query := mysql.DB.Model(&models.Book{})
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("isbn = ? OR title LIKE ? OR author LIKE ?", keyword, like, like)
	}
	writePage(c, query.Order("updated_at DESC"), page, pageSize, &[]models.Book{})
}

func (h *BookController) Save(c *gin.Context) {
	var req struct {
		Isbn        string `json:"isbn"`
		Title       string `json:"title"`
		Author      string `json:"author"`
		Publisher   string `json:"publisher"`
		PublishYear string `json:"publish_year"`
		CoverURL    string `json:"cover_url"`
		Source      string `json:"source"`
	}
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	isbn := normalizeISBNForAdmin(req.Isbn)
	if isbn == "" || strings.TrimSpace(req.Title) == "" {
		echo.Error(c, "Failed", "ISBN或书名无效")
		return
	}
	now := time.Now()
	book := models.Book{
		Isbn:        isbn,
		Title:       strings.TrimSpace(req.Title),
		Author:      req.Author,
		Publisher:   req.Publisher,
		PublishYear: req.PublishYear,
		CoverURL:    req.CoverURL,
		Source:      req.Source,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if book.Source == "" {
		book.Source = "manual"
	}
	err := mysql.DB.Transaction(func(tx *gorm.DB) error {
		var existing models.Book
		find := tx.Where("isbn = ?", isbn).First(&existing)
		if errors.Is(find.Error, gorm.ErrRecordNotFound) || find.RowsAffected == 0 {
			return tx.Create(&book).Error
		}
		return tx.Model(&existing).Updates(map[string]interface{}{
			"title": req.Title, "author": req.Author, "publisher": req.Publisher,
			"publish_year": req.PublishYear, "cover_url": req.CoverURL, "source": book.Source, "updated_at": now,
		}).Error
	})
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "book.save", "book", "SUCCESS", isbn)
	echo.Success(c, book, "")
}

func (h *PriceSnapshotController) List(c *gin.Context) {
	page, pageSize := pageParams(c)
	query := mysql.DB.Model(&models.PriceSnapshot{}).Select("isbn, source, min_price, avg_price, max_price, sample_count, confidence, raw_url, raw_payload_ref, collected_at, expires_at, created_at")
	if isbn := strings.TrimSpace(c.Query("isbn")); isbn != "" {
		query = query.Where("isbn = ?", normalizeISBNForAdmin(isbn))
	}
	writePage(c, query.Order("collected_at DESC"), page, pageSize, &[]models.PriceSnapshot{})
}

func (h *PriceSnapshotController) Create(c *gin.Context) {
	var req struct {
		Isbn        string `json:"isbn"`
		Source      string `json:"source"`
		MinPrice    string `json:"min_price"`
		AvgPrice    string `json:"avg_price"`
		MaxPrice    string `json:"max_price"`
		SampleCount int    `json:"sample_count"`
		Confidence  string `json:"confidence"`
		RawURL      string `json:"raw_url"`
	}
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	isbn := normalizeISBNForAdmin(req.Isbn)
	if isbn == "" {
		echo.Error(c, "Failed", "ISBN无效")
		return
	}
	minPrice, err := optionalDecimal(req.MinPrice)
	if err != nil {
		echo.Error(c, "Failed", "最低价格式错误")
		return
	}
	avgPrice, err := optionalDecimal(req.AvgPrice)
	if err != nil {
		echo.Error(c, "Failed", "平均价格式错误")
		return
	}
	maxPrice, err := optionalDecimal(req.MaxPrice)
	if err != nil {
		echo.Error(c, "Failed", "最高价格式错误")
		return
	}
	now := time.Now()
	if req.Source == "" {
		req.Source = "manual"
	}
	if req.Confidence == "" {
		req.Confidence = "NONE"
	}
	price := models.PriceSnapshot{
		Isbn:        isbn,
		Source:      req.Source,
		MinPrice:    minPrice,
		AvgPrice:    avgPrice,
		MaxPrice:    maxPrice,
		SampleCount: req.SampleCount,
		Confidence:  strings.ToUpper(req.Confidence),
		RawURL:      req.RawURL,
		CollectedAt: now,
		CreatedAt:   now,
	}
	if err := mysql.DB.Create(&price).Error; err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "price_snapshot.create", "price_snapshot", "SUCCESS", isbn)
	echo.Success(c, price, "")
}

func (h *RecycleRuleController) Current(c *gin.Context) {
	var rule models.RecycleRule
	tx := mysql.DB.Where("enabled = ?", true).Order("updated_at DESC").First(&rule)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) || tx.RowsAffected == 0 {
		echo.Success(c, nil, "")
		return
	}
	if tx.Error != nil {
		echo.Error(c, "Failed", tx.Error.Error())
		return
	}
	echo.Success(c, rule, "")
}

func (h *RecycleRuleController) SaveAndEnable(c *gin.Context) {
	var req struct {
		Name              string `json:"name"`
		MinAcceptAvgPrice string `json:"min_accept_avg_price"`
		RecycleRate       string `json:"recycle_rate"`
		MinSampleCount    int    `json:"min_sample_count"`
		PriceValidDays    int    `json:"price_valid_days"`
		LowConfidenceMode string `json:"low_confidence_mode"`
	}
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	minAvg, err := decimal.NewFromString(req.MinAcceptAvgPrice)
	if err != nil {
		echo.Error(c, "Failed", "最低均价格式错误")
		return
	}
	rate, err := decimal.NewFromString(req.RecycleRate)
	if err != nil {
		echo.Error(c, "Failed", "回收比例格式错误")
		return
	}
	now := time.Now()
	rule := models.RecycleRule{
		Version:           "rule-" + now.Format("20060102150405"),
		Name:              req.Name,
		MinAcceptAvgPrice: minAvg,
		RecycleRate:       rate,
		MinSampleCount:    req.MinSampleCount,
		PriceValidDays:    req.PriceValidDays,
		LowConfidenceMode: req.LowConfidenceMode,
		Enabled:           true,
		CreatedBy:         currentSession(c).UserID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if rule.Name == "" {
		rule.Name = "后台规则"
	}
	if rule.MinSampleCount <= 0 {
		rule.MinSampleCount = 3
	}
	if rule.PriceValidDays <= 0 {
		rule.PriceValidDays = 30
	}
	if rule.LowConfidenceMode == "" {
		rule.LowConfidenceMode = "NEED_REVIEW"
	}
	err = mysql.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.RecycleRule{}).Where("enabled = ?", true).Update("enabled", false).Error; err != nil {
			return err
		}
		return tx.Create(&rule).Error
	})
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "recycle_rule.save_and_enable", "recycle_rule", "SUCCESS", rule.ID)
	echo.Success(c, rule, "")
}

func (h *ImportController) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		echo.Error(c, "Failed", "请选择导入文件")
		return
	}
	result, err := parseImportFile(file)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	now := time.Now()
	session := currentSession(c)
	task := models.ImportTask{
		FileName:    file.Filename,
		FileType:    strings.TrimPrefix(strings.ToLower(filepath.Ext(file.Filename)), "."),
		Status:      "COMPLETED",
		TotalRows:   result.Total,
		SuccessRows: result.ValidCount,
		FailedRows:  result.InvalidCount,
		CreatedBy:   session.UserID,
		CreatedAt:   now,
		CompletedAt: &now,
	}
	err = mysql.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
		for _, input := range result.Rows {
			raw := input.RawPayload
			row := models.ImportTaskRow{
				TaskID:       task.ID,
				RowNumber:    input.RowNumber,
				Isbn:         input.Isbn,
				Title:        input.Title,
				Valid:        input.Valid,
				ErrorMessage: input.ErrorMessage,
				RawPayload:   raw,
				CreatedAt:    now,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			if input.Valid {
				book := models.Book{Isbn: input.Isbn, Title: input.Title, Author: input.Author, Publisher: input.Publisher, PublishYear: input.PublishYear, Source: input.Source, CreatedAt: now, UpdatedAt: now}
				if err := tx.Where(models.Book{Isbn: input.Isbn}).FirstOrCreate(&book).Error; err != nil {
					return err
				}
				if input.AvgPrice != nil || input.MinPrice != nil || input.MaxPrice != nil {
					price := models.PriceSnapshot{Isbn: input.Isbn, Source: input.Source, MinPrice: input.MinPrice, AvgPrice: input.AvgPrice, MaxPrice: input.MaxPrice, SampleCount: input.SampleCount, Confidence: input.Confidence, CollectedAt: now, CreatedAt: now}
					if err := tx.Create(&price).Error; err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "import.upload", "import_task", "SUCCESS", task.ID)
	echo.Success(c, gin.H{"task": task, "parse_result": result}, "")
}

func (h *ImportController) SyncKongfzCategory(c *gin.Context) {
	var req kongfzCategorySyncRequest
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = adminSvc.KongfzCategorySource
	}
	input := adminSvc.KongfzCategorySyncInput{
		CatID:     req.CatID,
		Source:    source,
		CreatedBy: currentSession(c).UserID,
		Items:     make([]adminSvc.KongfzCategoryItemInput, 0, len(req.Items)),
	}
	for index, item := range req.Items {
		rawPayload := ""
		if len(item.RawPayload) > 0 {
			rawPayload = string(item.RawPayload)
		}
		input.Items = append(input.Items, adminSvc.KongfzCategoryItemInput{
			RowNumber:        index + 1,
			Page:             item.Page,
			Isbn:             item.Isbn,
			Title:            item.Title,
			BookName:         item.BookName,
			Author:           item.Author,
			Publisher:        item.Publisher,
			PublishYear:      item.PublishYear,
			PublishDate:      item.PublishDate,
			Binding:          item.Binding,
			ListPrice:        item.ListPrice,
			CoverURL:         item.CoverURL,
			RawURL:           item.RawURL,
			KongfzID:         item.KongfzID,
			Mid:              item.Mid,
			OldBookMinPrice:  kongfzDecimalFromRequest(item.OldBookMinPrice, item.OldBookMinPriceText),
			OldBookOnSaleNum: item.OldBookOnSaleNum,
			BookShowInfo:     item.BookShowInfo,
			RawPayload:       rawPayload,
		})
	}
	result, err := adminSvc.SyncKongfzCategory(input)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "import.kongfz_category.sync", "import_task", "SUCCESS", result.Task.ID)
	echo.Success(c, result, "")
}

func (h *ImportController) List(c *gin.Context) {
	page, pageSize := pageParams(c)
	writePage(c, mysql.DB.Model(&models.ImportTask{}).Order("created_at DESC"), page, pageSize, &[]models.ImportTask{})
}

func (h *ImportController) Detail(c *gin.Context) {
	taskID := c.Query("id")
	var task models.ImportTask
	if err := mysql.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		echo.Error(c, "Failed", "导入任务不存在")
		return
	}
	var rows []models.ImportTaskRow
	if err := mysql.DB.Where("task_id = ?", taskID).Order("row_no ASC").Find(&rows).Error; err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	echo.Success(c, gin.H{"task": task, "rows": rows}, "")
}

func pageParams(c *gin.Context) (int, int) {
	page := parsePositiveInt(c.Query("page"), 1)
	pageSize := parsePositiveInt(c.Query("page_size"), 20)
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func writePage(c *gin.Context, query *gorm.DB, page, pageSize int, out interface{}) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(out).Error; err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	echo.Success(c, response.AdminPageResponse{Page: page, PageSize: pageSize, Total: total, Items: out}, "")
}

func optionalDecimal(value string) (*decimal.Decimal, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return nil, err
	}
	return &amount, nil
}

func kongfzDecimalFromRequest(value interface{}, text string) *decimal.Decimal {
	candidates := make([]string, 0, 2)
	switch typed := value.(type) {
	case float64:
		candidates = append(candidates, strconv.FormatFloat(typed, 'f', -1, 64))
	case string:
		candidates = append(candidates, typed)
	}
	if strings.TrimSpace(text) != "" {
		candidates = append(candidates, text)
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		amount, err := decimal.NewFromString(candidate)
		if err == nil {
			return &amount
		}
	}
	return nil
}

func parseImportFile(file *multipart.FileHeader) (adminSvc.ImportParseResult, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	src, err := file.Open()
	if err != nil {
		return adminSvc.ImportParseResult{}, err
	}
	defer src.Close()
	if ext == ".csv" {
		return adminSvc.ParseCSVImportRows(src)
	}
	if ext == ".xlsx" {
		return parseXLSXImportRows(src)
	}
	return adminSvc.ImportParseResult{}, errors.New("仅支持CSV或XLSX")
}

func parseXLSXImportRows(reader io.Reader) (adminSvc.ImportParseResult, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return adminSvc.ImportParseResult{}, err
	}
	defer f.Close()
	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return adminSvc.ImportParseResult{}, err
	}
	var builder strings.Builder
	for _, row := range rows {
		for i, col := range row {
			if i > 0 {
				builder.WriteByte(',')
			}
			encoded, _ := json.Marshal(col)
			builder.WriteString(string(encoded))
		}
		builder.WriteByte('\n')
	}
	return adminSvc.ParseCSVImportRows(strings.NewReader(builder.String()))
}

func normalizeISBNForAdmin(isbn string) string {
	return adminSvc.NormalizeISBN(isbn)
}
