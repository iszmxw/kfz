package v1

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"goapi/app/models"
	"goapi/app/response"
	adminSvc "goapi/app/services/admin"
	"goapi/pkg/config"
	"goapi/pkg/echo"
	"goapi/pkg/helpers"
	"goapi/pkg/mysql"
	"goapi/pkg/redis"
	"gorm.io/gorm"
)

type AuthController struct {
	BaseController
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthController) Login(c *gin.Context) {
	var req loginRequest
	if err := bindJSON(c, &req); err != nil {
		echo.Error(c, "Failed", "请求参数错误")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		echo.Error(c, "Failed", "用户名或密码不能为空")
		return
	}

	user, ok, err := findAdminUser(req.Username)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	if !ok {
		user, ok = fallbackConfigAdmin(req.Username, req.Password)
		if !ok {
			echo.Error(c, "Failed", "用户名或密码错误")
			return
		}
	} else if user.Status != "ACTIVE" || !adminSvc.CheckPassword(user.PasswordHash, req.Password) {
		echo.Error(c, "Failed", "用户名或密码错误")
		return
	}

	roles, permissions, err := loadUserRolesAndPermissions(user.ID)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	if len(permissions) == 0 && user.Username == config.GetString("admin.default_username", "admin") {
		permissions = []string{"*"}
		roles = []string{"SUPER_ADMIN"}
	}
	menus, err := loadMenusForPermissions(permissions)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}

	token := helpers.GetUUID()
	session := adminSvc.Session{
		UserID:      user.ID,
		Username:    user.Username,
		Name:        user.Name,
		Status:      user.Status,
		Roles:       roles,
		Permissions: permissions,
	}
	rawSession, _ := json.Marshal(session)
	if redis.Client == nil {
		echo.Error(c, "Failed", "Redis未初始化")
		return
	}
	if _, err := redis.Client.Add("Admin:Token:"+token, string(rawSession), config.GetInt("admin.token_ttl_seconds", 86400)); err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}

	now := time.Now()
	_ = mysql.DB.Model(&models.AdminUser{}).
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{"last_login_at": now, "last_login_ip": c.ClientIP(), "updated_at": now}).Error

	echo.Success(c, response.AdminLoginResponse{
		Token:       token,
		User:        session.UserDTO(),
		Permissions: permissions,
		Menus:       menus,
	}, "")
}

func (h *AuthController) Logout(c *gin.Context) {
	token := strings.TrimPrefix(strings.TrimSpace(c.GetHeader("Authorization")), "Bearer ")
	if token != "" && redis.Client != nil {
		redis.Client.Delete("Admin:Token:" + token)
	}
	echo.Success(c, gin.H{"ok": true}, "")
}

func (h *AuthController) Me(c *gin.Context) {
	session := currentSession(c)
	echo.Success(c, gin.H{
		"user":        session.UserDTO(),
		"permissions": session.Permissions,
	}, "")
}

func (h *AuthController) Menus(c *gin.Context) {
	session := currentSession(c)
	menus, err := loadMenusForPermissions(session.Permissions)
	if err != nil {
		echo.Error(c, "Failed", err.Error())
		return
	}
	echo.Success(c, menus, "")
}

func findAdminUser(username string) (models.AdminUser, bool, error) {
	var user models.AdminUser
	tx := mysql.DB.Where("username = ?", username).First(&user)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) || tx.RowsAffected == 0 {
		return user, false, nil
	}
	return user, tx.Error == nil, tx.Error
}

func fallbackConfigAdmin(username, password string) (models.AdminUser, bool) {
	defaultUsername := config.GetString("admin.default_username", "admin")
	defaultPassword := config.GetString("admin.default_password", "")
	if username != defaultUsername || defaultPassword == "" || password != defaultPassword {
		return models.AdminUser{}, false
	}
	return models.AdminUser{
		ID:       "config_admin",
		Username: defaultUsername,
		Name:     "超级管理员",
		Status:   "ACTIVE",
	}, true
}

func loadUserRolesAndPermissions(userID string) ([]string, []string, error) {
	var roles []models.AdminRole
	err := mysql.DB.Table((&models.AdminRole{}).TableName()+" AS r").
		Select("r.*").
		Joins("JOIN "+(&models.AdminUserRole{}).TableName()+" ur ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.status = ?", userID, "ACTIVE").
		Find(&roles).Error
	if err != nil {
		return nil, nil, err
	}

	roleIDs := make([]string, 0, len(roles))
	roleCodes := make([]string, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
		roleCodes = append(roleCodes, role.Code)
	}
	if len(roleIDs) == 0 {
		return roleCodes, nil, nil
	}

	var rolePermissions []models.AdminRolePermission
	if err := mysql.DB.Where("role_id IN ?", roleIDs).Find(&rolePermissions).Error; err != nil {
		return nil, nil, err
	}
	permissions := make([]string, 0, len(rolePermissions))
	seen := make(map[string]struct{})
	for _, permission := range rolePermissions {
		if _, ok := seen[permission.PermissionCode]; ok {
			continue
		}
		seen[permission.PermissionCode] = struct{}{}
		permissions = append(permissions, permission.PermissionCode)
	}
	return roleCodes, permissions, nil
}

func loadMenusForPermissions(permissions []string) ([]response.AdminMenuDTO, error) {
	var menus []models.AdminMenu
	if err := mysql.DB.Where("status = ?", "ACTIVE").Order("sort ASC").Find(&menus).Error; err != nil {
		return nil, err
	}
	set := adminSvc.NewPermissionSet(permissions)
	result := make([]response.AdminMenuDTO, 0, len(menus))
	for _, menu := range menus {
		if !set.Allows(menu.PermissionCode) {
			continue
		}
		result = append(result, response.AdminMenuDTO{
			ID:             menu.ID,
			ParentID:       menu.ParentID,
			Title:          menu.Title,
			Path:           menu.Path,
			Icon:           menu.Icon,
			PermissionCode: menu.PermissionCode,
		})
	}
	return result, nil
}
