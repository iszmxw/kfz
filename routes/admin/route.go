package admin

import (
	adminV1 "goapi/app/controllers/admin/v1"
	middlewaresV1 "goapi/app/middlewares/v1"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(router *gin.RouterGroup) {
	group := new(adminV1.Group)
	adminRoute := router.Group("/admin/api/v1")
	{
		auth := adminRoute.Group("/auth")
		{
			auth.POST("/login", group.AuthController.Login)
			auth.POST("/logout", middlewaresV1.AdminAuth(""), group.AuthController.Logout)
			auth.GET("/me", middlewaresV1.AdminAuth(""), group.AuthController.Me)
			auth.GET("/menus", middlewaresV1.AdminAuth(""), group.AuthController.Menus)
		}

		adminRoute.GET("/dashboard/summary", middlewaresV1.AdminAuth("dashboard:read"), group.DashboardController.Summary)

		manualReview := adminRoute.Group("/manual-review", middlewaresV1.AdminAuth("manual_review:read"))
		{
			manualReview.GET("/list", group.ManualReviewController.List)
			manualReview.POST("/decide", middlewaresV1.AdminAuth("manual_review:write"), group.ManualReviewController.Decide)
		}

		adminRoute.GET("/scan-log/list", middlewaresV1.AdminAuth("scan_log:read"), group.ScanLogController.List)

		book := adminRoute.Group("/book")
		{
			book.GET("/list", middlewaresV1.AdminAuth("book:read"), group.BookController.List)
			book.POST("/save", middlewaresV1.AdminAuth("book:write"), group.BookController.Save)
		}

		price := adminRoute.Group("/price-snapshot")
		{
			price.GET("/list", middlewaresV1.AdminAuth("price_snapshot:read"), group.PriceSnapshotController.List)
			price.POST("/create", middlewaresV1.AdminAuth("price_snapshot:write"), group.PriceSnapshotController.Create)
		}

		importTask := adminRoute.Group("/import")
		{
			importTask.POST("/upload", middlewaresV1.AdminAuth("import:write"), group.ImportController.Upload)
			importTask.POST("/kongfz-category/sync", middlewaresV1.AdminAuth("import:write"), group.ImportController.SyncKongfzCategory)
			importTask.GET("/list", middlewaresV1.AdminAuth("import:read"), group.ImportController.List)
			importTask.GET("/detail", middlewaresV1.AdminAuth("import:read"), group.ImportController.Detail)
		}

		rule := adminRoute.Group("/recycle-rule")
		{
			rule.GET("/current", middlewaresV1.AdminAuth("recycle_rule:read"), group.RecycleRuleController.Current)
			rule.POST("/save-and-enable", middlewaresV1.AdminAuth("recycle_rule:write"), group.RecycleRuleController.SaveAndEnable)
		}

		system := adminRoute.Group("/system", middlewaresV1.AdminAuth("system:read"))
		{
			system.GET("/user/list", group.SystemController.UserList)
			system.POST("/user/save", middlewaresV1.AdminAuth("system:write"), group.SystemController.UserSave)
			system.GET("/role/list", group.SystemController.RoleList)
			system.POST("/role/save", middlewaresV1.AdminAuth("system:write"), group.SystemController.RoleSave)
			system.POST("/role/assign-permissions", middlewaresV1.AdminAuth("system:write"), group.SystemController.RoleAssignPermissions)
			system.GET("/menu/list", group.SystemController.MenuList)
			system.POST("/menu/save", middlewaresV1.AdminAuth("system:write"), group.SystemController.MenuSave)
			system.GET("/api-permission/list", group.SystemController.APIPermissionList)
			system.POST("/api-permission/save", middlewaresV1.AdminAuth("system:write"), group.SystemController.APIPermissionSave)
			system.GET("/operation-log/list", group.SystemController.OperationLogList)
		}
	}
}
