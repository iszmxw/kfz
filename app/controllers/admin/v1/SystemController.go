package v1

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"goapi/app/models"
	"goapi/app/response"
	adminSvc "goapi/app/services/admin"
	"goapi/pkg/echo"
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
	var total int64
	if err := query.Count(&total).Error; err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	var users []models.AdminUser
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	items, err := buildAdminUserListItems(users)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	echo.Success(c, response.AdminPageResponse{Page: page, PageSize: pageSize, Total: total, Items: items}, "")
}

func (h *SystemController) UserSave(c *gin.Context) {
	var req struct {
		ID       uint64   `json:"id"`
		Username string   `json:"username"`
		Password string   `json:"password"`
		Name     string   `json:"name"`
		Status   string   `json:"status"`
		RoleIDs  []uint64 `json:"role_ids"`
	}
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Name = strings.TrimSpace(req.Name)
	req.Status = strings.TrimSpace(req.Status)
	if req.Username == "" || req.Name == "" {
		echo.Error(c, "Failed", "用户名或姓名不能为空")
		return
	}
	if req.ID == 0 && strings.TrimSpace(req.Password) == "" {
		echo.Error(c, "Failed", "创建用户时密码不能为空")
		return
	}
	now := time.Now()
	if req.Status == "" {
		req.Status = "ACTIVE"
	}
	err := mysql.DB.Transaction(func(tx *gorm.DB) error {
		userID := req.ID
		if userID == 0 {
			hash, err := adminSvc.HashPassword(req.Password)
			if err != nil {
				return err
			}
			user := models.AdminUser{Username: req.Username, PasswordHash: hash, Name: req.Name, Status: req.Status, CreatedAt: now, UpdatedAt: now}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
			userID = user.ID
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
			if roleID == 0 {
				continue
			}
			if err := tx.Create(&models.AdminUserRole{UserID: userID, RoleID: roleID, CreatedAt: now}).Error; err != nil {
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
	if req.ID == 0 {
		req.CreatedAt = now
	}
	req.UpdatedAt = now
	if err := mysql.DB.Save(&req).Error; err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	recordOperation(c, "system.role.save", "admin_role", "SUCCESS", req.Code)
	echo.Success(c, req, "")
}

func (h *SystemController) RoleAssignPermissions(c *gin.Context) {
	var req struct {
		RoleID      uint64   `json:"role_id"`
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
			if err := tx.Create(&models.AdminRolePermission{RoleID: req.RoleID, PermissionType: "API", PermissionCode: code, CreatedAt: now}).Error; err != nil {
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
	if req.ID == 0 {
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
	recordOperation(c, "system.menu.save", "admin_menu", "SUCCESS", req.PermissionCode)
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
	if req.ID == 0 {
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
	recordOperation(c, "system.api_permission.save", "admin_api_permission", "SUCCESS", req.PermissionCode)
	echo.Success(c, req, "")
}

func buildAdminUserListItems(users []models.AdminUser) ([]response.AdminUserListItem, error) {
	items := make([]response.AdminUserListItem, 0, len(users))
	if len(users) == 0 {
		return items, nil
	}
	userIDs := make([]uint64, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	var joins []models.AdminUserRole
	if err := mysql.DB.Where("user_id IN ?", userIDs).Find(&joins).Error; err != nil {
		return nil, err
	}
	roleIDs := make([]uint64, 0, len(joins))
	roleIDsByUser := make(map[uint64][]uint64)
	for _, join := range joins {
		roleIDsByUser[join.UserID] = append(roleIDsByUser[join.UserID], join.RoleID)
		roleIDs = append(roleIDs, join.RoleID)
	}

	rolesByID := map[uint64]models.AdminRole{}
	if len(roleIDs) > 0 {
		var roles []models.AdminRole
		if err := mysql.DB.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
			return nil, err
		}
		for _, role := range roles {
			rolesByID[role.ID] = role
		}
	}

	for _, user := range users {
		roleIDs := roleIDsByUser[user.ID]
		roleCodes := make([]string, 0, len(roleIDs))
		for _, roleID := range roleIDs {
			if role, ok := rolesByID[roleID]; ok {
				roleCodes = append(roleCodes, role.Code)
			}
		}
		var lastLoginAt *string
		if user.LastLoginAt != nil {
			formatted := formatTime(*user.LastLoginAt)
			lastLoginAt = &formatted
		}
		items = append(items, response.AdminUserListItem{
			ID:          user.ID,
			Username:    user.Username,
			Name:        user.Name,
			Status:      user.Status,
			RoleIDs:     roleIDs,
			Roles:       roleCodes,
			LastLoginAt: lastLoginAt,
			LastLoginIP: user.LastLoginIP,
			CreatedAt:   formatTime(user.CreatedAt),
			UpdatedAt:   formatTime(user.UpdatedAt),
		})
	}
	return items, nil
}

func (h *SystemController) OperationLogList(c *gin.Context) {
	page, pageSize := pageParams(c)
	writePage(c, mysql.DB.Model(&models.AdminOperationLog{}).Order("created_at DESC"), page, pageSize, &[]models.AdminOperationLog{})
}
