// Package main is the entry point for GPT-Load Manager.
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gin-gonic/gin"
)

//go:embed web/dist/*
var webFS embed.FS

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// API routes
	api := router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})
	}

	// Serve frontend
	staticFS, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		log.Printf("No embedded frontend, checking disk: %v", err)
		webDir := os.Getenv("WEB_DIR")
		if webDir == "" {
			webDir = filepath.Join(".", "web", "dist")
		}
		if _, err := os.Stat(webDir); err == nil {
			router.Static("/", webDir)
			router.NoRoute(func(c *gin.Context) {
				c.File(filepath.Join(webDir, "index.html"))
			})
		}
	} else {
		router.StaticFS("/", http.FS(staticFS))
		router.NoRoute(func(c *gin.Context) {
			c.FileFromFS("index.html", http.FS(staticFS))
		})
	}

	// Start server
	port := "3002"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("GPT-Load Manager starting on :%s", port)
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down...")
}
