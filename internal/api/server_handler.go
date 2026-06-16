package api

import (
	"net/http"
	"time"

	"github.com/solider245/gpt-load-manager/internal/db"
	"github.com/solider245/gpt-load-manager/internal/monitor"
	"github.com/solider245/gpt-load-manager/internal/ssh"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ServerHandler handles server CRUD and health check requests.
type ServerHandler struct {
	db     *gorm.DB
	ssh    *ssh.Connector
	poller *monitor.Poller
}

// NewServerHandler creates a new ServerHandler.
func NewServerHandler(database *gorm.DB, connector *ssh.Connector, poller *monitor.Poller) *ServerHandler {
	return &ServerHandler{db: database, ssh: connector, poller: poller}
}

// List returns all servers.
func (h *ServerHandler) List(c *gin.Context) {
	var servers []db.Server
	if err := h.db.Order("created_at desc").Find(&servers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": servers})
}

// Get returns a single server by ID.
func (h *ServerHandler) Get(c *gin.Context) {
	var server db.Server
	if err := h.db.First(&server, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": server})
}

// Create adds a new server to the inventory.
func (h *ServerHandler) Create(c *gin.Context) {
	var req db.Server
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	if req.Name == "" || req.Host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and host are required"})
		return
	}

	if err := h.db.Create(&req).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "server name already exists"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": req})
}

// Update modifies an existing server.
func (h *ServerHandler) Update(c *gin.Context) {
	var server db.Server
	if err := h.db.First(&server, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// prevent clearing critical fields
	if name, ok := updates["name"]; ok && name == "" {
		delete(updates, "name")
	}
	if host, ok := updates["host"]; ok && host == "" {
		delete(updates, "host")
	}

	updates["updated_at"] = time.Now()
	if err := h.db.Model(&server).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.db.First(&server, server.ID)
	c.JSON(http.StatusOK, gin.H{"data": server})
}

// Delete removes a server.
func (h *ServerHandler) Delete(c *gin.Context) {
	if err := h.db.Delete(&db.Server{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// CheckHealth performs health checks (SSH + HTTP) on the server.
func (h *ServerHandler) CheckHealth(c *gin.Context) {
	var server db.Server
	if err := h.db.First(&server, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	result := gin.H{
		"ssh":    nil,
		"http":   nil,
		"online": false,
	}

	// SSH connectivity check
	sshInfo, sshErr := h.ssh.TestConnection(server.Host, server.SSHPort, server.AuthType, server.AuthCredential)
	if sshErr != nil {
		result["ssh"] = gin.H{"online": false, "error": sshErr.Error()}
	} else {
		result["ssh"] = gin.H{"online": true, "info": sshInfo}
	}

	// HTTP health check
	healthResult := h.poller.HealthCheck(server.Host, server.GPTPort)
	result["http"] = healthResult

	if healthResult.Online {
		result["online"] = true
	}

	// Update server status in database
	status := "offline"
	if healthResult.Online {
		status = "online"
	}
	now := time.Now()
	h.db.Model(&server).Updates(map[string]interface{}{
		"status":        status,
		"last_health_at": &now,
		"version":       healthResult.Uptime, // placeholder
	})

	result["server_id"] = server.ID
	c.JSON(http.StatusOK, gin.H{"data": result})
}
