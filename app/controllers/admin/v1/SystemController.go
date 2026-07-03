package v1

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"goapi/app/models"
	adminSvc "goapi/app/services/admin"
	"goapi/pkg/echo"
	"goapi/pkg/helpers"
	"goapi/pkg/mysql"
	"gorm.io/gorm"
)

type SystemController struct {
	BaseController
}

func (h *SystemController) UserList(c *gin.Context) {
	page, pageSize := pageParams(c)
	query := mysql.DB.Model(&models.AdminUser{})
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		query = query.Where("username LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	writePage(c, query.Order("created_at DESC"), page, pageSize, &[]models.AdminUser{})
}

func (h *SystemController) UserSave(c *gin.Context) {
	var req struct {
		ID       string   `json:"id"`
		Username string   `json:"username"`
		Password string   `json:"password"`
		Name     string   `json:"name"`
		Status   string   `json:"status"`
		RoleIDs  []string `json:"role_ids"`
	}
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Name) == "" {
		echo.Error(c, "Failed", "用户名或姓名不能为空")
		return
	}
	now := time.Now()
	if req.Status == "" {
		req.Status = "ACTIVE"
	}
	err := mysql.DB.Transaction(func(tx *gorm.DB) error {
		userID := req.ID
		if userID == "" {
			hash, err := adminSvc.HashPassword(req.Password)
			if err != nil {
				return err
			}
			userID = helpers.GetUUID()
			if err := tx.Create(&models.AdminUser{ID: userID, Username: req.Username, PasswordHash: hash, Name: req.Name, Status: req.Status, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
				return err
			}
		} else {
			updates := map[string]interface{}{"name": req.Name, "status": req.Status, "updated_at": now}
			if req.Password != "" {
				hash, err := adminSvc.HashPassword(req.Password)
				if err != nil {
					return err
				}
				updates["password_hash"] = hash
			}
			if err := tx.Model(&models.AdminUser{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
				return err
			}
			if err := tx.Where("user_id = ?", userID).Delete(&models.AdminUserRole{}).Error; err != nil {
				return err
			}
		}
		for _, roleID := range req.RoleIDs {
			if roleID == "" {
				continue
			}
			if err := tx.Create(&models.AdminUserRole{ID: helpers.GetUUID(), UserID: userID, RoleID: roleID, CreatedAt: now}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "system.user.save", "admin_user", "SUCCESS", req.Username)
	echo.Success(c, gin.H{"ok": true}, "")
}

func (h *SystemController) RoleList(c *gin.Context) {
	page, pageSize := pageParams(c)
	writePage(c, mysql.DB.Model(&models.AdminRole{}).Order("created_at DESC"), page, pageSize, &[]models.AdminRole{})
}

func (h *SystemController) RoleSave(c *gin.Context) {
	var req models.AdminRole
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	now := time.Now()
	if req.Status == "" {
		req.Status = "ACTIVE"
	}
	if req.ID == "" {
		req.ID = helpers.GetUUID()
		req.CreatedAt = now
	}
	req.UpdatedAt = now
	if err := mysql.DB.Save(&req).Error; err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "system.role.save", "admin_role", "SUCCESS", req.ID)
	echo.Success(c, req, "")
}

func (h *SystemController) RoleAssignPermissions(c *gin.Context) {
	var req struct {
		RoleID      string   `json:"role_id"`
		Permissions []string `json:"permissions"`
	}
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	now := time.Now()
	err := mysql.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", req.RoleID).Delete(&models.AdminRolePermission{}).Error; err != nil {
			return err
		}
		for _, code := range req.Permissions {
			code = strings.TrimSpace(code)
			if code == "" {
				continue
			}
			if err := tx.Create(&models.AdminRolePermission{ID: helpers.GetUUID(), RoleID: req.RoleID, PermissionType: "API", PermissionCode: code, CreatedAt: now}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "system.role.assign_permissions", "admin_role_permission", "SUCCESS", req.RoleID)
	echo.Success(c, gin.H{"ok": true}, "")
}

func (h *SystemController) MenuList(c *gin.Context) {
	page, pageSize := pageParams(c)
	writePage(c, mysql.DB.Model(&models.AdminMenu{}).Order("sort ASC"), page, pageSize, &[]models.AdminMenu{})
}

func (h *SystemController) MenuSave(c *gin.Context) {
	var req models.AdminMenu
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	now := time.Now()
	if req.ID == "" {
		req.ID = helpers.GetUUID()
		req.CreatedAt = now
	}
	if req.Status == "" {
		req.Status = "ACTIVE"
	}
	req.UpdatedAt = now
	if err := mysql.DB.Save(&req).Error; err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "system.menu.save", "admin_menu", "SUCCESS", req.ID)
	echo.Success(c, req, "")
}

func (h *SystemController) APIPermissionList(c *gin.Context) {
	page, pageSize := pageParams(c)
	writePage(c, mysql.DB.Model(&models.AdminAPIPermission{}).Order("created_at DESC"), page, pageSize, &[]models.AdminAPIPermission{})
}

func (h *SystemController) APIPermissionSave(c *gin.Context) {
	var req models.AdminAPIPermission
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	now := time.Now()
	if req.ID == "" {
		req.ID = helpers.GetUUID()
		req.CreatedAt = now
	}
	if req.Status == "" {
		req.Status = "ACTIVE"
	}
	req.Method = strings.ToUpper(req.Method)
	req.UpdatedAt = now
	if err := mysql.DB.Save(&req).Error; err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "system.api_permission.save", "admin_api_permission", "SUCCESS", req.ID)
	echo.Success(c, req, "")
}

func (h *SystemController) OperationLogList(c *gin.Context) {
	page, pageSize := pageParams(c)
	writePage(c, mysql.DB.Model(&models.AdminOperationLog{}).Order("created_at DESC"), page, pageSize, &[]models.AdminOperationLog{})
}
