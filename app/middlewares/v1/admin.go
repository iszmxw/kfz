package v1

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	adminSvc "goapi/app/services/admin"
	"goapi/pkg/echo"
	"goapi/pkg/redis"
)

func AdminAuth(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader("Authorization"))
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" {
			echo.Error(c, "LoginInvalid", "")
			c.Abort()
			return
		}
		if redis.Client == nil {
			echo.Error(c, "Failed", "Redis未初始化")
			c.Abort()
			return
		}
		raw, err := redis.Client.Get("Admin:Token:" + token)
		if err != nil || raw == "" {
			echo.Error(c, "LoginInvalid", "")
			c.Abort()
			return
		}
		var session adminSvc.Session
		if err := json.Unmarshal([]byte(raw), &session); err != nil {
			echo.Error(c, "LoginInvalid", "")
			c.Abort()
			return
		}
		if !adminSvc.NewPermissionSet(session.Permissions).Allows(permissionCode) {
			echo.Error(c, "Failed", "无权限")
			c.Abort()
			return
		}
		c.Set("admin_session", session)
		c.Set("admin_user", session.UserDTO())
		c.Next()
	}
}
