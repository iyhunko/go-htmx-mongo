package integration

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/iyhunko/go-htmx-mongo/internal/model"
	"github.com/iyhunko/go-htmx-mongo/internal/repository"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var testDB *mongo.Database

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

func TestIntegrationMongoPostRepository_Create(t *testing.T) {
	pool, resource, db := setupMongoDB(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	repo := repository.NewMongoPostRepository(db)
	ctx := context.Background()

	post := model.NewPost("Test Title", "Test Content")

	err := repo.Create(ctx, post)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if post.ID.IsZero() {
		t.Errorf("Create() should set post ID")
	}

	if post.CreatedAt.IsZero() {
		t.Errorf("Create() should set CreatedAt")
	}

	if post.UpdatedAt.IsZero() {
		t.Errorf("Create() should set UpdatedAt")
	}
}

func TestIntegrationMongoPostRepository_FindByID(t *testing.T) {
	pool, resource, db := setupMongoDB(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	repo := repository.NewMongoPostRepository(db)
	ctx := context.Background()

	// Create a post
	post := model.NewPost("Test Title", "Test Content")
	if err := repo.Create(ctx, post); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Find the post
	found, err := repo.FindByID(ctx, post.ID.Hex())
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.ID != post.ID {
		t.Errorf("FindByID() ID = %v, want %v", found.ID, post.ID)
	}

	if found.Title != post.Title {
		t.Errorf("FindByID() Title = %v, want %v", found.Title, post.Title)
	}

	if found.Content != post.Content {
		t.Errorf("FindByID() Content = %v, want %v", found.Content, post.Content)
	}
}

func TestIntegrationMongoPostRepository_FindByID_NotFound(t *testing.T) {
	pool, resource, db := setupMongoDB(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	repo := repository.NewMongoPostRepository(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "507f1f77bcf86cd799439011")
	if err != repository.ErrPostNotFound {
		t.Errorf("FindByID() error = %v, want %v", err, repository.ErrPostNotFound)
	}
}

func TestIntegrationMongoPostRepository_FindAll(t *testing.T) {
	pool, resource, db := setupMongoDB(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	repo := repository.NewMongoPostRepository(db)
	ctx := context.Background()

	// Create multiple posts
	for i := 1; i <= 15; i++ {
		post := model.NewPost(fmt.Sprintf("Title %d", i), fmt.Sprintf("Content %d", i))
		if err := repo.Create(ctx, post); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Test pagination
	posts, err := repo.FindAll(ctx, 10, 0)
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}

	if len(posts) != 10 {
		t.Errorf("FindAll() returned %d posts, want 10", len(posts))
	}

	// Test second page
	posts2, err := repo.FindAll(ctx, 10, 10)
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}

	if len(posts2) != 5 {
		t.Errorf("FindAll() returned %d posts, want 5", len(posts2))
	}
}

func TestIntegrationMongoPostRepository_Search(t *testing.T) {
	pool, resource, db := setupMongoDB(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	repo := repository.NewMongoPostRepository(db)
	ctx := context.Background()

	// Create posts with different content
	testCases := []struct {
		title   string
		content string
	}{
		{"Go Programming", "Learn Go language"},
		{"Python Basics", "Introduction to Python"},
		{"Go Advanced", "Advanced Go concepts"},
	}

	for _, tc := range testCases {
		post := model.NewPost(tc.title, tc.content)
		if err := repo.Create(ctx, post); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Search for "Go"
	posts, err := repo.Search(ctx, "Go", 10, 0)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(posts) != 2 {
		t.Errorf("Search() returned %d posts, want 2", len(posts))
	}

	// Search for "Python"
	posts, err = repo.Search(ctx, "Python", 10, 0)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("Search() returned %d posts, want 1", len(posts))
	}
}

func TestIntegrationMongoPostRepository_Update(t *testing.T) {
	pool, resource, db := setupMongoDB(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	repo := repository.NewMongoPostRepository(db)
	ctx := context.Background()

	// Create a post
	post := model.NewPost("Original Title", "Original Content")
	if err := repo.Create(ctx, post); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update the post
	post.Update("Updated Title", "Updated Content")
	if err := repo.Update(ctx, post); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update
	found, err := repo.FindByID(ctx, post.ID.Hex())
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.Title != "Updated Title" {
		t.Errorf("Update() Title = %v, want Updated Title", found.Title)
	}

	if found.Content != "Updated Content" {
		t.Errorf("Update() Content = %v, want Updated Content", found.Content)
	}
}

func TestIntegrationMongoPostRepository_Delete(t *testing.T) {
	pool, resource, db := setupMongoDB(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	repo := repository.NewMongoPostRepository(db)
	ctx := context.Background()

	// Create a post
	post := model.NewPost("Test Title", "Test Content")
	if err := repo.Create(ctx, post); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Delete the post
	if err := repo.Delete(ctx, post.ID.Hex()); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err := repo.FindByID(ctx, post.ID.Hex())
	if err != repository.ErrPostNotFound {
		t.Errorf("Delete() post still exists")
	}
}

func TestIntegrationMongoPostRepository_Count(t *testing.T) {
	pool, resource, db := setupMongoDB(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	repo := repository.NewMongoPostRepository(db)
	ctx := context.Background()

	// Create posts
	for i := 1; i <= 5; i++ {
		post := model.NewPost(fmt.Sprintf("Title %d", i), fmt.Sprintf("Content %d", i))
		if err := repo.Create(ctx, post); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}

	if count != 5 {
		t.Errorf("Count() = %d, want 5", count)
	}
}

func TestIntegrationMongoPostRepository_CountSearch(t *testing.T) {
	pool, resource, db := setupMongoDB(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	repo := repository.NewMongoPostRepository(db)
	ctx := context.Background()

	// Create posts
	testCases := []struct {
		title   string
		content string
	}{
		{"Go Programming", "Learn Go language"},
		{"Python Basics", "Introduction to Python"},
		{"Go Advanced", "Advanced Go concepts"},
	}

	for _, tc := range testCases {
		post := model.NewPost(tc.title, tc.content)
		if err := repo.Create(ctx, post); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	count, err := repo.CountSearch(ctx, "Go")
	if err != nil {
		t.Fatalf("CountSearch() error = %v", err)
	}

	if count != 2 {
		t.Errorf("CountSearch() = %d, want 2", count)
	}
}
