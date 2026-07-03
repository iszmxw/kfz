package v1

import (
	"time"

	"github.com/gin-gonic/gin"
	"goapi/app/models"
	"goapi/app/response"
	"goapi/pkg/echo"
	"goapi/pkg/mysql"
)

type DashboardController struct {
	BaseController
}

func (h *DashboardController) Summary(c *gin.Context) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var todayScanCount, pendingReviewCount, acceptedTodayCount, lowConfidenceCount int64
	var suggestedTotal float64
	_ = mysql.DB.Model(&models.ScanLog{}).Where("scanned_at >= ?", start).Count(&todayScanCount).Error
	_ = mysql.DB.Model(&models.ScanLog{}).
		Where("decision = ?", "NEED_REVIEW").
		Where("NOT EXISTS (SELECT 1 FROM " + (&models.ManualDecision{}).TableName() + " md WHERE md.scan_log_id = " + (&models.ScanLog{}).TableName() + ".id)").
		Count(&pendingReviewCount).Error
	_ = mysql.DB.Model(&models.ScanLog{}).Where("scanned_at >= ? AND decision = ?", start, "ACCEPT").Count(&acceptedTodayCount).Error
	_ = mysql.DB.Model(&models.ScanLog{}).Where("confidence IN ?", []string{"LOW", "NONE"}).Count(&lowConfidenceCount).Error
	_ = mysql.DB.Model(&models.ScanLog{}).Where("scanned_at >= ? AND suggested_recycle_price IS NOT NULL", start).Select("COALESCE(SUM(suggested_recycle_price),0)").Scan(&suggestedTotal).Error

	echo.Success(c, response.DashboardSummary{
		TodayScanCount:      todayScanCount,
		PendingReviewCount:  pendingReviewCount,
		AcceptedTodayCount:  acceptedTodayCount,
		SuggestedTotalPrice: suggestedTotal,
		LowConfidenceCount:  lowConfidenceCount,
	}, "")
}
