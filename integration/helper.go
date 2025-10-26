package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iyhunko/go-htmx-mongo/internal/controller"
	httproutes "github.com/iyhunko/go-htmx-mongo/internal/http"
	"github.com/iyhunko/go-htmx-mongo/internal/repository"
	"github.com/iyhunko/go-htmx-mongo/internal/service"
	"github.com/iyhunko/go-htmx-mongo/pkg/config"
	"github.com/iyhunko/go-htmx-mongo/web"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// findProjectRoot finds the project root directory
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find project root")
}

// setupMongoDB creates a test MongoDB instance using dockertest
func setupMongoDB(t *testing.T) (*dockertest.Pool, *dockertest.Resource, *mongo.Database) {
	// Skip if in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create dockertest pool
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not construct pool: %s", err)
	}

	// Set max wait time
	pool.MaxWait = 120 * time.Second

	// Pull mongodb image
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "7",
		Env: []string{
			"MONGO_INITDB_DATABASE=testdb",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	var client *mongo.Client
	var db *mongo.Database

	// Exponential backoff-retry
	if err := pool.Retry(func() error {
		var err error
		mongoURI := fmt.Sprintf("mongodb://localhost:%s", resource.GetPort("27017/tcp"))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
		if err != nil {
			return err
		}

		// Ping the database
		if err := client.Ping(ctx, nil); err != nil {
			return err
		}

		db = client.Database("testdb")
		return nil
	}); err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	return pool, resource, db
}

// setupTestServer creates a test HTTP server with all dependencies
func setupTestServer(t *testing.T) (*gin.Engine, *dockertest.Pool, *dockertest.Resource, *mongo.Database) {
	// Skip if in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Change to project root for template loading
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}
	defer os.Chdir(originalDir)

	// Setup MongoDB
	pool, resource, db := setupMongoDB(t)

	// Initialize application layers
	postRepo := repository.NewMongoPostRepository(db)
	postService := service.NewPostService(postRepo)

	// Load templates
	templates, err := web.LoadTemplates()
	if err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

	// Create config
	cfg := &config.Config{
		PageSizeLimit: 10,
	}

	postController := controller.NewPostController(postService, templates, cfg)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	httproutes.SetupRoutes(router, postController)

	return router, pool, resource, db
}
