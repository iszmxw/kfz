package client

import (
	clientV1 "goapi/app/controllers/client/v1"
	middlewaresV1 "goapi/app/middlewares/v1"

	"github.com/gin-gonic/gin"
)

// RegisterClientRoutes 注册Client路由
func RegisterClientRoutes(router *gin.RouterGroup) {
	//router.Use(middlewaresV1.AddressLimit()) // 地区限制
	router.Use(middlewaresV1.Client())
	// 路由分组 客户端 模块
	AppRoute := router.Group("/app")
	{
		{ // V1 版本
			clientV1Group := new(clientV1.Group)
			V1Route := AppRoute.Group("/v1")

			// Demo 接口
			demo := V1Route.Group("/demo")
			{
				demo.GET("/ping.json", clientV1Group.DemoController.Ping)
			}

			// Book 接口
			book := V1Route.Group("/book")
			{
				book.GET("/check.json", clientV1Group.BookController.Check)
				book.GET("/detail.json", clientV1Group.BookController.Detail)
			}

			// Scan 接口
			scan := V1Route.Group("/scan")
			{
				scan.GET("/history.json", clientV1Group.ScanController.History)
			}
		}

	}
}
