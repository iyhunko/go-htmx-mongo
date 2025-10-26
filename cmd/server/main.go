package main

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iyhunko/go-htmx-mongo/internal/controller"
	"github.com/iyhunko/go-htmx-mongo/internal/db"
	httproutes "github.com/iyhunko/go-htmx-mongo/internal/http"
	"github.com/iyhunko/go-htmx-mongo/internal/http/middleware"
	"github.com/iyhunko/go-htmx-mongo/internal/repository"
	"github.com/iyhunko/go-htmx-mongo/internal/service"
	"github.com/iyhunko/go-htmx-mongo/pkg/config"
	"github.com/iyhunko/go-htmx-mongo/web"
)

func main() {
	initLogger()
	slog.Info("Starting application")

	cfg := config.Load()
	slog.Info("Configuration loaded", "serverPort", cfg.HttpServerPort, "database", cfg.MongoDBDatabase)

	mongodb := connectDatabase(cfg)
	defer disconnectDatabase(mongodb)

	postController := initializeController(mongodb, cfg)
	router := setupRouter(postController)
	server := createServer(cfg, router)

	startServer(server, cfg)
	waitForShutdown(server)
}

// initLogger initializes the structured logger
func initLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

// connectDatabase establishes a connection to MongoDB
func connectDatabase(cfg *config.Config) *db.MongoDB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongodb, err := db.Connect(ctx, cfg.GetMongoURI(), cfg.MongoDBDatabase)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	return mongodb
}

// disconnectDatabase closes the database connection
func disconnectDatabase(mongodb *db.MongoDB) {
	if err := mongodb.Disconnect(context.Background()); err != nil {
		slog.Error("Error disconnecting from MongoDB", "error", err)
	}
}

// initializeController initializes all application layers
func initializeController(mongodb *db.MongoDB, cfg *config.Config) *controller.PostController {
	postRepo := repository.NewMongoPostRepository(mongodb.DB)
	postService := service.NewPostService(postRepo)

	templates := loadTemplates()
	return controller.NewPostController(postService, templates, cfg)
}

// loadTemplates loads HTML templates
func loadTemplates() *template.Template {
	templates, err := web.LoadTemplates()
	if err != nil {
		slog.Error("Failed to load templates", "error", err)
		os.Exit(1)
	}
	slog.Info("Templates loaded successfully")
	return templates
}

// setupRouter configures the Gin router with middleware and routes
func setupRouter(postController *controller.PostController) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	httproutes.SetupRoutes(router, postController)
	return router
}

// createServer creates an HTTP server with appropriate timeouts
func createServer(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// startServer starts the HTTP server in a goroutine
func startServer(server *http.Server, cfg *config.Config) {
	go func() {
		slog.Info("Server starting", "port", cfg.HttpServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()
}

// waitForShutdown waits for interrupt signal and performs graceful shutdown
func waitForShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped gracefully")
}
