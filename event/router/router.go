package router

import (
	"github.com/gin-gonic/gin"
	"go-ethereum-demo/event/global"
	"go-ethereum-demo/event/service"
	"net/http"
)

func RegisterRouter(r *gin.Engine) {
	rootGroup := r.Group(global.Config.Server.ContextPath)
	rootGroup.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": "success",
			"message": "pong",
		})
	})

	planGroup := rootGroup.Group("/plan")
	{
		planGroup.POST("/create", service.CreateEventParsePlan)
	}
}
