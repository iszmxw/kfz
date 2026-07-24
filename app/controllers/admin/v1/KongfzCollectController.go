package v1

import (
	"strings"

	"github.com/gin-gonic/gin"
	"goapi/app/models"
	adminSvc "goapi/app/services/admin"
	"goapi/pkg/echo"
	"goapi/pkg/mysql"
)

type KongfzCollectController struct{ BaseController }

func (h *KongfzCollectController) Create(c *gin.Context) {
	var req struct {
		CatID      int    `json:"cat_id"`
		StartPage  int    `json:"start_page"`
		EndPage    int    `json:"end_page"`
		DelayMS    int    `json:"delay_ms"`
		RetryTimes int    `json:"retry_times"`
		UserAgent  string `json:"user_agent"`
		Cookie     string `json:"cookie"`
	}
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	task, err := adminSvc.CreateKongfzCollectTask(adminSvc.KongfzCollectCreateInput{
		CatID:      req.CatID,
		StartPage:  req.StartPage,
		EndPage:    req.EndPage,
		DelayMS:    req.DelayMS,
		RetryTimes: req.RetryTimes,
		UserAgent:  req.UserAgent,
		Cookie:     req.Cookie,
		CreatedBy:  currentSession(c).UserID,
	})
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "kongfz_collect.task.create", "kongfz_collect_task", "SUCCESS", task.ID)
	echo.Success(c, task, "")
}

func (h *KongfzCollectController) List(c *gin.Context) {
	if !adminSvc.KongfzCollectSchemaReady() {
		echo.Error(c, "Failed", "孔夫子采集表不存在，请先执行 migrations/001_init.sql")
		return
	}
	page, pageSize := pageParams(c)
	query := mysql.DB.Model(&models.KongfzCollectTask{})
	if status := strings.ToUpper(strings.TrimSpace(c.Query("status"))); status != "" {
		query = query.Where("status = ?", status)
	}
	writePage(c, query.Order("created_at DESC"), page, pageSize, &[]models.KongfzCollectTask{})
}

func (h *KongfzCollectController) Detail(c *gin.Context) {
	taskID := parsePositiveUint(c.Query("id"))
	result, err := adminSvc.GetKongfzCollectTaskDetail(taskID, parsePositiveInt(c.Query("row_limit"), 50))
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	echo.Success(c, result, "")
}

func (h *KongfzCollectController) Stop(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id"`
	}
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	if err := adminSvc.StopKongfzCollectTask(req.ID); err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "kongfz_collect.task.stop", "kongfz_collect_task", "SUCCESS", req.ID)
	echo.Success(c, gin.H{"ok": true}, "")
}

func (h *KongfzCollectController) Retry(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id"`
	}
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	task, err := adminSvc.RetryKongfzCollectTask(adminSvc.KongfzCollectRetryInput{TaskID: req.ID, CreatedBy: currentSession(c).UserID})
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "kongfz_collect.task.retry", "kongfz_collect_task", "SUCCESS", task.ID)
	echo.Success(c, task, "")
}

func parsePositiveUint(value string) uint64 {
	parsed := parsePositiveInt(value, 0)
	if parsed < 0 {
		return 0
	}
	return uint64(parsed)
}
