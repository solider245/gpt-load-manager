package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/solider245/gpt-load-manager/internal/api"
	"github.com/solider245/gpt-load-manager/internal/config"
	"github.com/solider245/gpt-load-manager/internal/db"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	database, err := db.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database initialized:", cfg.DBPath)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	api.RegisterRoutes(router.Group("/api"), database)

	serveFrontend(router, cfg.WebDir)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("GPT-Load Manager starting on :%s", cfg.Port)
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down...")
}

func serveFrontend(router *gin.Engine, webDir string) {
	if webDir == "" {
		webDir = filepath.Join(".", "web", "dist")
	}
	if _, err := os.Stat(filepath.Join(webDir, "index.html")); err != nil {
		log.Printf("Frontend not found at %s (API-only mode): %v", webDir, err)
		return
	}

	// Serve compiled static assets
	assetsDir := filepath.Join(webDir, "assets")
	if fi, err := os.Stat(assetsDir); err == nil && fi.IsDir() {
		router.Static("/assets", assetsDir)
	}

	// SPA fallback: serve index.html for all non-API paths
	router.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}
		c.File(filepath.Join(webDir, "index.html"))
	})

	log.Println("Frontend:", webDir)
}
