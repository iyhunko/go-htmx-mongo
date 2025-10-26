package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	httproutes "github.com/iyhunko/go-htmx-mongo/http"
	"github.com/iyhunko/go-htmx-mongo/internal/controller"
	"github.com/iyhunko/go-htmx-mongo/internal/db"
	"github.com/iyhunko/go-htmx-mongo/internal/repository"
	"github.com/iyhunko/go-htmx-mongo/internal/service"
	"github.com/iyhunko/go-htmx-mongo/pkg/config"
	"github.com/iyhunko/go-htmx-mongo/web"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting application")

	// Load configuration
	cfg := config.Load()
	slog.Info("Configuration loaded", "serverPort", cfg.ServerPort, "database", cfg.MongoDB)

	// Connect to MongoDB with auto-migration
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongodb, err := db.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := mongodb.Disconnect(context.Background()); err != nil {
			slog.Error("Error disconnecting from MongoDB", "error", err)
		}
	}()

	// Initialize layers
	postRepo := repository.NewMongoPostRepository(mongodb.DB)
	postService := service.NewPostService(postRepo)

	// Load templates with custom functions
	templates, err := web.LoadTemplates()
	if err != nil {
		slog.Error("Failed to load templates", "error", err)
		os.Exit(1)
	}
	slog.Info("Templates loaded successfully")

	postController := controller.NewPostController(postService, templates)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Custom logger middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start)
		slog.Info("Request processed",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"duration", duration.String(),
		)
	})

	// Setup routes
	httproutes.SetupRoutes(router, postController)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Server starting", "port", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped gracefully")
}
