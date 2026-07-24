package admin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"goapi/app/models"
	"goapi/pkg/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	KongfzCollectStatusPending       = "PENDING"
	KongfzCollectStatusRunning       = "RUNNING"
	KongfzCollectStatusCompleted     = "COMPLETED"
	KongfzCollectStatusFailed        = "FAILED"
	KongfzCollectStatusStopRequested = "STOP_REQUESTED"
	KongfzCollectStatusStopped       = "STOPPED"

	KongfzRawSyncStatusPending  = "PENDING"
	KongfzRawSyncStatusImported = "IMPORTED"
	KongfzRawSyncStatusFailed   = "FAILED"
)

const (
	defaultKongfzCatID      = 43
	defaultKongfzStartPage  = 1
	defaultKongfzDelayMS    = 1200
	defaultKongfzRetryTimes = 2
	maxKongfzAutoPages      = 100
)

type KongfzCollectCreateInput struct {
	CatID      int
	StartPage  int
	EndPage    int
	DelayMS    int
	RetryTimes int
	UserAgent  string
	Cookie     string
	CreatedBy  uint64
}

type KongfzCollectRetryInput struct {
	TaskID    uint64
	CreatedBy uint64
}

type KongfzCollectTaskDetail struct {
	Task models.KongfzCollectTask     `json:"task"`
	Rows []models.KongfzCollectRawRow `json:"rows"`
}

type kongfzRequestHeaders struct {
	UserAgent string `json:"user_agent,omitempty"`
	Cookie    string `json:"cookie,omitempty"`
}

type kongfzPageResponse struct {
	Status  int    `json:"status"`
	ErrType string `json:"errType"`
	Message string `json:"message"`
	Data    struct {
		ItemResponse struct {
			Total int               `json:"total"`
			List  []json.RawMessage `json:"list"`
			Pager kongfzPager       `json:"pager"`
		} `json:"itemResponse"`
	} `json:"data"`
}

type kongfzPager struct {
	Page  int `json:"page"`
	Size  int `json:"size"`
	Total int `json:"total"`
	Pages int `json:"pages"`
}

type mappedKongfzCollectItem struct {
	Input          KongfzCategoryItemInput
	Isbn           string
	NormalizedIsbn string
	Title          string
	Valid          bool
	ErrorMessage   string
	RawHash        string
	RawPayload     string
}

type KongfzCollector struct {
	client *http.Client
}

func NewKongfzCollector(client *http.Client) *KongfzCollector {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &KongfzCollector{client: client}
}

func CreateKongfzCollectTask(input KongfzCollectCreateInput) (models.KongfzCollectTask, error) {
	if !KongfzCollectSchemaReady() {
		return models.KongfzCollectTask{}, errors.New("孔夫子采集表不存在，请先执行 migrations/001_init.sql")
	}
	if !KongfzCoreSchemaReady() {
		return models.KongfzCollectTask{}, errors.New("核心业务表结构不是最终版本，请重建本地库并执行 migrations/001_init.sql")
	}
	if !KongfzImportSchemaReady() {
		return models.KongfzCollectTask{}, errors.New("导入任务表结构不是最终版本，请重建本地库并执行 migrations/001_init.sql")
	}
	now := time.Now()
	input = normalizeKongfzCollectInput(input)
	headers, err := json.Marshal(kongfzRequestHeaders{UserAgent: strings.TrimSpace(input.UserAgent), Cookie: strings.TrimSpace(input.Cookie)})
	if err != nil {
		return models.KongfzCollectTask{}, err
	}
	task := models.KongfzCollectTask{
		CatID:          input.CatID,
		StartPage:      input.StartPage,
		EndPage:        input.EndPage,
		DelayMS:        input.DelayMS,
		RetryTimes:     input.RetryTimes,
		Status:         KongfzCollectStatusPending,
		CurrentPage:    0,
		RequestHeaders: string(headers),
		CreatedBy:      input.CreatedBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := mysql.DB.Create(&task).Error; err != nil {
		return models.KongfzCollectTask{}, err
	}
	return task, nil
}

func StopKongfzCollectTask(taskID uint64) error {
	if !KongfzCollectSchemaReady() {
		return errors.New("孔夫子采集表不存在，请先执行 migrations/001_init.sql")
	}
	if taskID == 0 {
		return errors.New("任务ID无效")
	}
	now := time.Now()
	tx := mysql.DB.Model(&models.KongfzCollectTask{}).
		Where("id = ? AND status IN ?", taskID, []string{KongfzCollectStatusPending, KongfzCollectStatusRunning}).
		Updates(map[string]interface{}{"status": KongfzCollectStatusStopRequested, "updated_at": now})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("任务不存在或不可停止")
	}
	return nil
}

func RetryKongfzCollectTask(input KongfzCollectRetryInput) (models.KongfzCollectTask, error) {
	if !KongfzCollectSchemaReady() {
		return models.KongfzCollectTask{}, errors.New("孔夫子采集表不存在，请先执行 migrations/001_init.sql")
	}
	if !KongfzCoreSchemaReady() {
		return models.KongfzCollectTask{}, errors.New("核心业务表结构不是最终版本，请重建本地库并执行 migrations/001_init.sql")
	}
	if !KongfzImportSchemaReady() {
		return models.KongfzCollectTask{}, errors.New("导入任务表结构不是最终版本，请重建本地库并执行 migrations/001_init.sql")
	}
	if input.TaskID == 0 {
		return models.KongfzCollectTask{}, errors.New("任务ID无效")
	}
	var old models.KongfzCollectTask
	tx := mysql.DB.Where("id = ?", input.TaskID).First(&old)
	if tx.Error != nil || tx.RowsAffected == 0 {
		return models.KongfzCollectTask{}, errors.New("任务不存在")
	}
	if old.Status != KongfzCollectStatusFailed && old.Status != KongfzCollectStatusStopped {
		return models.KongfzCollectTask{}, errors.New("仅失败或已停止任务可重试")
	}
	headers := parseKongfzRequestHeaders(old.RequestHeaders)
	return CreateKongfzCollectTask(KongfzCollectCreateInput{
		CatID:      old.CatID,
		StartPage:  old.StartPage,
		EndPage:    old.EndPage,
		DelayMS:    old.DelayMS,
		RetryTimes: old.RetryTimes,
		UserAgent:  headers.UserAgent,
		Cookie:     headers.Cookie,
		CreatedBy:  input.CreatedBy,
	})
}

func GetKongfzCollectTaskDetail(taskID uint64, rowLimit int) (KongfzCollectTaskDetail, error) {
	if !KongfzCollectSchemaReady() {
		return KongfzCollectTaskDetail{}, errors.New("孔夫子采集表不存在，请先执行 migrations/001_init.sql")
	}
	if taskID == 0 {
		return KongfzCollectTaskDetail{}, errors.New("任务ID无效")
	}
	if rowLimit <= 0 || rowLimit > 200 {
		rowLimit = 50
	}
	var task models.KongfzCollectTask
	tx := mysql.DB.Where("id = ?", taskID).First(&task)
	if tx.Error != nil || tx.RowsAffected == 0 {
		return KongfzCollectTaskDetail{}, errors.New("任务不存在")
	}
	var rows []models.KongfzCollectRawRow
	if err := mysql.DB.Where("task_id = ?", taskID).Order("page DESC, row_index DESC").Limit(rowLimit).Find(&rows).Error; err != nil {
		return KongfzCollectTaskDetail{}, err
	}
	return KongfzCollectTaskDetail{Task: task, Rows: rows}, nil
}

func (c *KongfzCollector) RunNextPendingTask(ctx context.Context) (bool, error) {
	task, ok, err := claimNextKongfzCollectTask()
	if err != nil || !ok {
		return ok, err
	}
	return true, c.RunTask(ctx, task.ID)
}

func (c *KongfzCollector) RunTask(ctx context.Context, taskID uint64) error {
	if !KongfzCoreSchemaReady() {
		err := errors.New("核心业务表结构不是最终版本，请重建本地库并执行 migrations/001_init.sql")
		_ = finishKongfzCollectTask(taskID, KongfzCollectStatusFailed, err.Error())
		return err
	}
	if !KongfzImportSchemaReady() {
		err := errors.New("导入任务表结构不是最终版本，请重建本地库并执行 migrations/001_init.sql")
		_ = finishKongfzCollectTask(taskID, KongfzCollectStatusFailed, err.Error())
		return err
	}
	var task models.KongfzCollectTask
	tx := mysql.DB.Where("id = ?", taskID).First(&task)
	if tx.Error != nil || tx.RowsAffected == 0 {
		return errors.New("任务不存在")
	}
	headers := parseKongfzRequestHeaders(task.RequestHeaders)
	totalPages := task.EndPage
	importedCount := task.ImportedCount
	collectedCount := task.CollectedCount
	validCount := task.ValidCount
	invalidCount := task.InvalidCount

	for page := task.StartPage; ; page++ {
		if err := ctx.Err(); err != nil {
			_ = finishKongfzCollectTask(task.ID, KongfzCollectStatusStopped, "采集已停止")
			return err
		}
		stop, err := kongfzStopRequested(task.ID)
		if err != nil {
			_ = finishKongfzCollectTask(task.ID, KongfzCollectStatusFailed, err.Error())
			return err
		}
		if stop {
			_ = finishKongfzCollectTask(task.ID, KongfzCollectStatusStopped, "采集已停止")
			return nil
		}
		if task.EndPage > 0 && page > task.EndPage {
			break
		}
		if task.EndPage == 0 && totalPages > 0 && page > totalPages {
			break
		}

		_ = updateKongfzCollectProgress(task.ID, map[string]interface{}{"current_page": page})
		payload, err := c.fetchPageWithRetry(ctx, task.CatID, page, task.RetryTimes, headers)
		if err != nil {
			_ = finishKongfzCollectTask(task.ID, KongfzCollectStatusFailed, err.Error())
			return err
		}
		if totalPages == 0 {
			totalPages = payload.Data.ItemResponse.Pager.Pages
			if totalPages <= 0 {
				totalPages = page
			}
			if totalPages > maxKongfzAutoPages {
				totalPages = maxKongfzAutoPages
			}
			_ = updateKongfzCollectProgress(task.ID, map[string]interface{}{"total_pages": totalPages})
		}
		if len(payload.Data.ItemResponse.List) == 0 {
			err := fmt.Errorf("第 %d 页为空，采集已停止", page)
			_ = finishKongfzCollectTask(task.ID, KongfzCollectStatusFailed, err.Error())
			return err
		}

		mapped, err := mapKongfzCollectPage(task.CatID, page, payload.Data.ItemResponse.List)
		if err != nil {
			_ = finishKongfzCollectTask(task.ID, KongfzCollectStatusFailed, err.Error())
			return err
		}
		imported, err := saveKongfzCollectPage(task, mapped)
		if err != nil {
			_ = finishKongfzCollectTask(task.ID, KongfzCollectStatusFailed, err.Error())
			return err
		}
		collectedCount += len(mapped)
		for _, item := range mapped {
			if item.Valid {
				validCount++
			} else {
				invalidCount++
			}
		}
		importedCount += imported
		_ = updateKongfzCollectProgress(task.ID, map[string]interface{}{
			"collected_count": collectedCount,
			"valid_count":     validCount,
			"invalid_count":   invalidCount,
			"imported_count":  importedCount,
		})

		if task.EndPage > 0 && page >= task.EndPage {
			break
		}
		if task.EndPage == 0 && totalPages > 0 && page >= totalPages {
			break
		}
		if task.DelayMS > 0 {
			select {
			case <-ctx.Done():
				_ = finishKongfzCollectTask(task.ID, KongfzCollectStatusStopped, "采集已停止")
				return ctx.Err()
			case <-time.After(time.Duration(task.DelayMS) * time.Millisecond):
			}
		}
	}
	return finishKongfzCollectTask(task.ID, KongfzCollectStatusCompleted, "")
}

func (c *KongfzCollector) FetchPage(ctx context.Context, catID, page int, headers kongfzRequestHeaders) (kongfzPageResponse, error) {
	requestURL := buildKongfzCategoryURL(catID, page)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return kongfzPageResponse{}, err
	}
	req.Header.Set("Accept", "application/json,text/plain,*/*")
	userAgent := strings.TrimSpace(headers.UserAgent)
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36"
	}
	req.Header.Set("User-Agent", userAgent)
	if cookie := strings.TrimSpace(headers.Cookie); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return kongfzPageResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		return kongfzPageResponse{}, fmt.Errorf("孔夫子接口返回 %d，采集已停止", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return kongfzPageResponse{}, fmt.Errorf("孔夫子接口返回 %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return kongfzPageResponse{}, err
	}
	if looksKongfzBlocked(string(body)) {
		return kongfzPageResponse{}, errors.New("孔夫子返回验证或风控页面，采集已停止")
	}
	var payload kongfzPageResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return kongfzPageResponse{}, errors.New("孔夫子接口返回非JSON内容，采集已停止")
	}
	if payload.Status != 1 {
		if payload.Message != "" {
			return kongfzPageResponse{}, errors.New(payload.Message)
		}
		return kongfzPageResponse{}, errors.New("孔夫子接口返回失败")
	}
	if payload.Data.ItemResponse.List == nil {
		return kongfzPageResponse{}, errors.New("孔夫子接口数据结构异常")
	}
	return payload, nil
}

func (c *KongfzCollector) fetchPageWithRetry(ctx context.Context, catID, page, retryTimes int, headers kongfzRequestHeaders) (kongfzPageResponse, error) {
	var lastErr error
	for attempt := 0; attempt <= retryTimes; attempt++ {
		payload, err := c.FetchPage(ctx, catID, page, headers)
		if err == nil {
			return payload, nil
		}
		lastErr = err
		if strings.Contains(err.Error(), "采集已停止") {
			return kongfzPageResponse{}, err
		}
		if attempt < retryTimes {
			select {
			case <-ctx.Done():
				return kongfzPageResponse{}, ctx.Err()
			case <-time.After(time.Duration(500*(attempt+1)) * time.Millisecond):
			}
		}
	}
	return kongfzPageResponse{}, lastErr
}

func mapKongfzCollectPage(catID, page int, rows []json.RawMessage) ([]mappedKongfzCollectItem, error) {
	mapped := make([]mappedKongfzCollectItem, 0, len(rows))
	for index, raw := range rows {
		item, err := mapKongfzCollectItem(catID, page, index, raw)
		if err != nil {
			return nil, err
		}
		mapped = append(mapped, item)
	}
	return mapped, nil
}

func mapKongfzCollectItem(catID, page, index int, raw json.RawMessage) (mappedKongfzCollectItem, error) {
	var record map[string]interface{}
	if err := json.Unmarshal(raw, &record); err != nil {
		return mappedKongfzCollectItem{}, err
	}
	bookShowInfo := stringSlice(record["bookShowInfo"])
	title := stringValue(record["bookName"])
	isbn := stringValue(record["isbn"])
	normalizedISBN := NormalizeISBN(isbn)
	publishDate := ""
	if len(bookShowInfo) > 2 {
		publishDate = bookShowInfo[2]
	}
	var rowErrors []string
	if normalizedISBN == "" {
		rowErrors = append(rowErrors, "ISBN格式无效")
	}
	if title == "" {
		rowErrors = append(rowErrors, "书名不能为空")
	}
	rawPayload := string(raw)
	rawHash := sha256Hex(rawPayload)
	input := KongfzCategoryItemInput{
		RowNumber:        index + 1,
		Page:             page,
		Isbn:             normalizedISBN,
		Title:            title,
		BookName:         title,
		Author:           valueAt(bookShowInfo, 0),
		Publisher:        valueAt(bookShowInfo, 1),
		PublishYear:      extractPublishYear(publishDate),
		PublishDate:      publishDate,
		Binding:          valueAt(bookShowInfo, 3),
		ListPrice:        valueAt(bookShowInfo, 4),
		CoverURL:         imageURL(record["imgUrlEntity"]),
		RawURL:           buildKongfzCategoryURL(catID, page),
		KongfzID:         uint64Value(record["id"]),
		Mid:              uint64Value(record["mid"]),
		OldBookMinPrice:  decimalFromInterface(record["oldBookMinPrice"]),
		OldBookOnSaleNum: intValue(record["oldBookOnSaleNum"]),
		BookShowInfo:     bookShowInfo,
		RawPayload:       rawPayload,
	}
	if input.OldBookMinPrice == nil {
		input.OldBookMinPrice = decimalFromInterface(record["oldBookMinPriceText"])
	}
	return mappedKongfzCollectItem{
		Input:          input,
		Isbn:           isbn,
		NormalizedIsbn: normalizedISBN,
		Title:          title,
		Valid:          len(rowErrors) == 0,
		ErrorMessage:   strings.Join(rowErrors, "；"),
		RawHash:        rawHash,
		RawPayload:     rawPayload,
	}, nil
}

func saveKongfzCollectPage(task models.KongfzCollectTask, mapped []mappedKongfzCollectItem) (int, error) {
	validInputs := make([]KongfzCategoryItemInput, 0, len(mapped))
	rawRows := make([]models.KongfzCollectRawRow, 0, len(mapped))
	now := time.Now()
	for index, item := range mapped {
		syncStatus := KongfzRawSyncStatusFailed
		if item.Valid {
			syncStatus = KongfzRawSyncStatusPending
			validInputs = append(validInputs, item.Input)
		}
		rawRows = append(rawRows, models.KongfzCollectRawRow{
			TaskID:         task.ID,
			CatID:          task.CatID,
			Page:           item.Input.Page,
			RowIndex:       index,
			Isbn:           item.Isbn,
			NormalizedIsbn: item.NormalizedIsbn,
			Title:          item.Title,
			CoverURL:       item.Input.CoverURL,
			Valid:          item.Valid,
			ErrorMessage:   item.ErrorMessage,
			RawHash:        item.RawHash,
			RawPayload:     item.RawPayload,
			SyncStatus:     syncStatus,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}
	if len(rawRows) == 0 {
		return 0, nil
	}
	err := mysql.DB.Transaction(func(tx *gorm.DB) error {
		for _, row := range rawRows {
			if row.NormalizedIsbn != "" {
				updates := map[string]interface{}{
					"task_id":            row.TaskID,
					"page":               row.Page,
					"row_index":          row.RowIndex,
					"isbn":               row.Isbn,
					"title":              row.Title,
					"cover_url":          row.CoverURL,
					"valid":              row.Valid,
					"error_message":      row.ErrorMessage,
					"raw_hash":           row.RawHash,
					"raw_payload":        row.RawPayload,
					"sync_status":        row.SyncStatus,
					"updated_at":         row.UpdatedAt,
					"import_task_row_id": 0,
				}
				update := tx.Model(&models.KongfzCollectRawRow{}).
					Where("cat_id = ? AND normalized_isbn = ?", row.CatID, row.NormalizedIsbn).
					Updates(updates)
				if update.Error != nil {
					return update.Error
				}
				if update.RowsAffected > 0 {
					continue
				}
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "task_id"}, {Name: "page"}, {Name: "row_index"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"isbn", "normalized_isbn", "title", "cover_url", "valid", "error_message", "raw_hash", "raw_payload", "sync_status", "updated_at",
				}),
			}).Create(&row).Error; err != nil {
				return err
			}
		}
		if len(validInputs) == 0 {
			return nil
		}
		result, err := SyncKongfzCategoryWithDB(tx, KongfzCategorySyncInput{
			CatID:     task.CatID,
			Source:    KongfzCategorySource,
			CreatedBy: task.CreatedBy,
			Items:     validInputs,
		})
		if err != nil {
			return err
		}
		rowByNumber := map[int]KongfzCategorySyncRow{}
		for _, row := range result.Rows {
			rowByNumber[row.RowNumber] = row
		}
		for _, item := range mapped {
			if !item.Valid {
				continue
			}
			syncRow := rowByNumber[item.Input.RowNumber]
			status := KongfzRawSyncStatusImported
			if !syncRow.Valid {
				status = KongfzRawSyncStatusFailed
			}
			if err := tx.Model(&models.KongfzCollectRawRow{}).
				Where("task_id = ? AND page = ? AND row_index = ?", task.ID, item.Input.Page, item.Input.RowNumber-1).
				Updates(map[string]interface{}{
					"sync_status":        status,
					"error_message":      syncRow.ErrorMessage,
					"import_task_row_id": syncRow.ImportTaskRowID,
					"updated_at":         now,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(validInputs), nil
}

func claimNextKongfzCollectTask() (models.KongfzCollectTask, bool, error) {
	if !KongfzCollectSchemaReady() {
		return models.KongfzCollectTask{}, false, nil
	}
	var task models.KongfzCollectTask
	now := time.Now()
	err := mysql.DB.Transaction(func(tx *gorm.DB) error {
		find := tx.Where("status = ?", KongfzCollectStatusPending).Order("created_at ASC").First(&task)
		if find.Error != nil {
			return find.Error
		}
		if find.RowsAffected == 0 || task.ID == 0 {
			return nil
		}
		return tx.Model(&models.KongfzCollectTask{}).Where("id = ? AND status = ?", task.ID, KongfzCollectStatusPending).
			Updates(map[string]interface{}{"status": KongfzCollectStatusRunning, "started_at": now, "updated_at": now}).Error
	})
	if err != nil {
		return models.KongfzCollectTask{}, false, err
	}
	if task.ID == 0 {
		return models.KongfzCollectTask{}, false, nil
	}
	task.Status = KongfzCollectStatusRunning
	task.StartedAt = &now
	return task, true, nil
}

func updateKongfzCollectProgress(taskID uint64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return mysql.DB.Model(&models.KongfzCollectTask{}).Where("id = ?", taskID).Updates(updates).Error
}

func finishKongfzCollectTask(taskID uint64, status, message string) error {
	now := time.Now()
	return mysql.DB.Model(&models.KongfzCollectTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":        status,
		"error_message": message,
		"completed_at":  now,
		"updated_at":    now,
	}).Error
}

func kongfzStopRequested(taskID uint64) (bool, error) {
	var task models.KongfzCollectTask
	tx := mysql.DB.Select("status").Where("id = ?", taskID).First(&task)
	if tx.Error != nil {
		return false, tx.Error
	}
	return task.Status == KongfzCollectStatusStopRequested, nil
}

func normalizeKongfzCollectInput(input KongfzCollectCreateInput) KongfzCollectCreateInput {
	if input.CatID <= 0 {
		input.CatID = defaultKongfzCatID
	}
	if input.StartPage <= 0 {
		input.StartPage = defaultKongfzStartPage
	}
	if input.EndPage > 0 && input.EndPage < input.StartPage {
		input.EndPage = input.StartPage
	}
	if input.DelayMS <= 0 {
		input.DelayMS = defaultKongfzDelayMS
	}
	if input.RetryTimes < 0 {
		input.RetryTimes = defaultKongfzRetryTimes
	}
	return input
}

func buildKongfzCategoryURL(catID, page int) string {
	values := url.Values{}
	values.Set("catId", strconv.Itoa(catID))
	values.Set("page", strconv.Itoa(page))
	return "https://search.kongfz.com/pc-gw/search-web/client/pc/bookLib/category/list?" + values.Encode()
}

func parseKongfzRequestHeaders(raw string) kongfzRequestHeaders {
	var headers kongfzRequestHeaders
	_ = json.Unmarshal([]byte(raw), &headers)
	return headers
}

func looksKongfzBlocked(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "captcha") || strings.Contains(text, "验证码") || strings.Contains(text, "访问过于频繁")
}

func stringSlice(value interface{}) []string {
	raw, ok := value.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, entry := range raw {
		result = append(result, strings.TrimSpace(fmt.Sprint(entry)))
	}
	return result
}

func stringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func intValue(value interface{}) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case string:
		parsed, _ := strconv.Atoi(strings.TrimSpace(typed))
		return parsed
	default:
		parsed, _ := strconv.Atoi(strings.TrimSpace(fmt.Sprint(value)))
		return parsed
	}
}

func uint64Value(value interface{}) uint64 {
	v := intValue(value)
	if v < 0 {
		return 0
	}
	return uint64(v)
}

func decimalFromInterface(value interface{}) *decimal.Decimal {
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" || text == "<nil>" {
		return nil
	}
	amount, err := decimal.NewFromString(text)
	if err != nil {
		return nil
	}
	return &amount
}

func imageURL(value interface{}) string {
	entity, ok := value.(map[string]interface{})
	if !ok {
		return ""
	}
	if big := stringValue(entity["bigImgUrl"]); big != "" && big != "<nil>" {
		return big
	}
	if small := stringValue(entity["smallImgUrl"]); small != "" && small != "<nil>" {
		return small
	}
	return ""
}

func valueAt(values []string, index int) string {
	if index >= len(values) {
		return ""
	}
	return values[index]
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func KongfzCollectSchemaReady() bool {
	if mysql.DB == nil {
		return false
	}
	return mysql.DB.Migrator().HasTable(&models.KongfzCollectTask{}) &&
		mysql.DB.Migrator().HasTable(&models.KongfzCollectRawRow{})
}

func KongfzImportSchemaReady() bool {
	if mysql.DB == nil {
		return false
	}
	migrator := mysql.DB.Migrator()
	return migrator.HasTable(&models.ImportTask{}) &&
		hasAutoIncrementBigintID(&models.ImportTask{}) &&
		migrator.HasTable(&models.ImportTaskRow{}) &&
		hasAutoIncrementBigintID(&models.ImportTaskRow{}) &&
		migrator.HasColumn(&models.ImportTaskRow{}, "task_id")
}

func KongfzCoreSchemaReady() bool {
	if mysql.DB == nil {
		return false
	}
	migrator := mysql.DB.Migrator()
	return migrator.HasTable(&models.Book{}) &&
		hasAutoIncrementBigintID(&models.Book{}) &&
		migrator.HasColumn(&models.Book{}, models.BookColumns.Isbn) &&
		migrator.HasColumn(&models.Book{}, models.BookColumns.Title) &&
		migrator.HasTable(&models.PriceSnapshot{}) &&
		hasAutoIncrementBigintID(&models.PriceSnapshot{}) &&
		migrator.HasColumn(&models.PriceSnapshot{}, models.PriceSnapshotColumns.Isbn) &&
		migrator.HasColumn(&models.PriceSnapshot{}, models.PriceSnapshotColumns.MinPrice) &&
		migrator.HasTable(&models.ScanLog{}) &&
		hasAutoIncrementBigintID(&models.ScanLog{}) &&
		migrator.HasColumn(&models.ScanLog{}, models.ScanLogColumns.PriceSnapshotID)
}

func hasAutoIncrementBigintID(model interface{}) bool {
	if mysql.DB.Dialector.Name() != "mysql" {
		return mysql.DB.Migrator().HasColumn(model, "id")
	}
	columns, err := mysql.DB.Migrator().ColumnTypes(model)
	if err != nil {
		return false
	}
	for _, column := range columns {
		if !strings.EqualFold(column.Name(), "id") {
			continue
		}
		dbType := strings.ToLower(column.DatabaseTypeName())
		autoIncrement, ok := column.AutoIncrement()
		return strings.Contains(dbType, "bigint") && (!ok || autoIncrement)
	}
	return false
}

var (
	kongfzWorkerOnce sync.Once
	kongfzWorkerStop context.CancelFunc
)

func StartKongfzCollectWorker() {
	kongfzWorkerOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		kongfzWorkerStop = cancel
		collector := NewKongfzCollector(nil)
		go func() {
			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()
			for {
				_, _ = collector.RunNextPendingTask(ctx)
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
			}
		}()
	})
}

func StopKongfzCollectWorker() {
	if kongfzWorkerStop != nil {
		kongfzWorkerStop()
	}
}
