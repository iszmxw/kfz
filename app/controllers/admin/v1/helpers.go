package v1

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"goapi/app/models"
	"goapi/app/response"
	adminSvc "goapi/app/services/admin"
	"goapi/pkg/helpers"
	"goapi/pkg/mysql"
)

func currentAdmin(c *gin.Context) response.AdminUserDTO {
	value, ok := c.Get("admin_user")
	if !ok {
		return response.AdminUserDTO{}
	}
	user, _ := value.(response.AdminUserDTO)
	return user
}

func currentSession(c *gin.Context) adminSvc.Session {
	value, ok := c.Get("admin_session")
	if !ok {
		return adminSvc.Session{}
	}
	session, _ := value.(adminSvc.Session)
	return session
}

func parsePositiveInt(value string, fallback int) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	result := 0
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

func decimalToFloatPtr(value *decimal.Decimal) *float64 {
	if value == nil {
		return nil
	}
	f, _ := value.Float64()
	result := helpers.Decimal(f, 2)
	return &result
}

func decimalToFloat(value decimal.Decimal) float64 {
	f, _ := value.Float64()
	return helpers.Decimal(f, 4)
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func bookDTO(book models.Book) response.BookDTO {
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

func bindJSON(c *gin.Context, target interface{}) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func recordOperation(c *gin.Context, action, resource, result string, requestBrief any) {
	session := currentSession(c)
	if session.UserID == 0 || mysql.DB == nil {
		return
	}
	now := time.Now()
	_ = mysql.DB.Create(&models.AdminOperationLog{
		UserID:       session.UserID,
		Username:     session.Username,
		Action:       action,
		Resource:     resource,
		RequestBrief: fmt.Sprint(requestBrief),
		Result:       result,
		IP:           c.ClientIP(),
		CreatedAt:    now,
	}).Error
}
