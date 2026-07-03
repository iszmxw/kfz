package admin

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"goapi/app/models"
	"goapi/pkg/helpers"
	"goapi/pkg/mysql"
	"gorm.io/gorm"
)

type ManualDecisionInput struct {
	ScanLogID          string
	AdminUserID        string
	ManualDecision     string
	ActualRecyclePrice *decimal.Decimal
	Note               string
}

func CreateManualDecision(input ManualDecisionInput) (models.ManualDecision, error) {
	if input.ManualDecision != "ACCEPT" && input.ManualDecision != "REJECT" {
		return models.ManualDecision{}, errors.New("人工确认结果无效")
	}

	var scan models.ScanLog
	tx := mysql.DB.Where(models.ScanLogColumns.ID+" = ?", input.ScanLogID).First(&scan)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) || tx.RowsAffected == 0 {
		return models.ManualDecision{}, errors.New("扫码记录不存在")
	}
	if tx.Error != nil {
		return models.ManualDecision{}, tx.Error
	}

	now := time.Now()
	decision := models.ManualDecision{
		ID:                 helpers.GetUUID(),
		ScanLogID:          scan.ID,
		Isbn:               scan.NormalizedIsbn,
		AdminUserID:        input.AdminUserID,
		ManualDecision:     input.ManualDecision,
		ActualRecyclePrice: input.ActualRecyclePrice,
		Note:               input.Note,
		CreatedAt:          now,
	}

	err := mysql.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(models.ManualDecision{ScanLogID: scan.ID}).FirstOrCreate(&decision).Error; err != nil {
			return err
		}
		return nil
	})
	return decision, err
}
