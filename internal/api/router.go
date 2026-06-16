// Package api registers all REST API routes.
package api

import (
	"github.com/solider245/gpt-load-manager/internal/monitor"
	"github.com/solider245/gpt-load-manager/internal/ssh"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes sets up all API routes on the given router group.
func RegisterRoutes(rg *gin.RouterGroup, database *gorm.DB) {
	sshConnector := ssh.NewConnector()
	healthPoller := monitor.NewPoller()
	handler := NewServerHandler(database, sshConnector, healthPoller)

	servers := rg.Group("/servers")
	{
		servers.GET("", handler.List)
		servers.POST("", handler.Create)
		servers.GET("/:id", handler.Get)
		servers.PUT("/:id", handler.Update)
		servers.DELETE("/:id", handler.Delete)
		servers.POST("/:id/check", handler.CheckHealth)
	}
}
