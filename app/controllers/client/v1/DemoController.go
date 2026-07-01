package v1

import (
	"github.com/gin-gonic/gin"
	"goapi/pkg/echo"
)

type DemoController struct {
	BaseController
}

func (h *DemoController) Ping(c *gin.Context) {
	echo.Success(c, gin.H{"message": "pong"}, "")
}
