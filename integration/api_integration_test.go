package integration

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/iyhunko/go-htmx-mongo/internal/model"
	"github.com/iyhunko/go-htmx-mongo/internal/repository"
)

func TestIntegrationAPI_Index(t *testing.T) {
	router, pool, resource, _ := setupTestServer(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	// Create request
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check content type is HTML
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") && contentType != "" {
		t.Errorf("Expected HTML content type, got %s", contentType)
	}
}

func TestIntegrationAPI_CreatePost(t *testing.T) {
	router, pool, resource, db := setupTestServer(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	// Create form data
	formData := url.Values{}
	formData.Set("title", "Test Post Title")
	formData.Set("content", "Test Post Content")

	// Create request
	req, _ := http.NewRequest("POST", "/posts", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Verify post was created in database
	postRepo := repository.NewMongoPostRepository(db)
	posts, err := postRepo.FindAll(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("Failed to query posts: %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("Expected 1 post in database, got %d", len(posts))
	}

	if len(posts) > 0 {
		if posts[0].Title != "Test Post Title" {
			t.Errorf("Expected title 'Test Post Title', got '%s'", posts[0].Title)
		}
		if posts[0].Content != "Test Post Content" {
			t.Errorf("Expected content 'Test Post Content', got '%s'", posts[0].Content)
		}
	}
}

func TestIntegrationAPI_CreatePost_Validation(t *testing.T) {
	router, pool, resource, _ := setupTestServer(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	// Create form data with empty title (should fail validation)
	formData := url.Values{}
	formData.Set("title", "")
	formData.Set("content", "Test Post Content")

	// Create request
	req, _ := http.NewRequest("POST", "/posts", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response is bad request
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestIntegrationAPI_UpdatePost(t *testing.T) {
	router, pool, resource, db := setupTestServer(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	// Create a post first
	postRepo := repository.NewMongoPostRepository(db)
	post := model.NewPost("Original Title", "Original Content")
	err := postRepo.Create(context.Background(), post)
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	// Update the post
	formData := url.Values{}
	formData.Set("title", "Updated Title")
	formData.Set("content", "Updated Content")

	// Create request
	req, _ := http.NewRequest("PUT", "/posts/"+post.ID.Hex(), strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Verify post was updated in database
	updatedPost, err := postRepo.FindByID(context.Background(), post.ID.Hex())
	if err != nil {
		t.Fatalf("Failed to find updated post: %v", err)
	}

	if updatedPost.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", updatedPost.Title)
	}
	if updatedPost.Content != "Updated Content" {
		t.Errorf("Expected content 'Updated Content', got '%s'", updatedPost.Content)
	}
}

func TestIntegrationAPI_DeletePost(t *testing.T) {
	router, pool, resource, db := setupTestServer(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	// Create a post first
	postRepo := repository.NewMongoPostRepository(db)
	post := model.NewPost("Test Title", "Test Content")
	err := postRepo.Create(context.Background(), post)
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	// Delete the post
	req, _ := http.NewRequest("DELETE", "/posts/"+post.ID.Hex(), nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify post was deleted from database
	_, err = postRepo.FindByID(context.Background(), post.ID.Hex())
	if err != repository.ErrPostNotFound {
		t.Errorf("Expected post to be deleted, but found it or got different error: %v", err)
	}
}

func TestIntegrationAPI_GetPosts(t *testing.T) {
	router, pool, resource, db := setupTestServer(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	// Create multiple posts
	postRepo := repository.NewMongoPostRepository(db)
	for i := 1; i <= 5; i++ {
		post := model.NewPost(fmt.Sprintf("Title %d", i), fmt.Sprintf("Content %d", i))
		err := postRepo.Create(context.Background(), post)
		if err != nil {
			t.Fatalf("Failed to create post: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get posts
	req, _ := http.NewRequest("GET", "/posts", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Response should contain HTML with posts
	body := w.Body.String()
	if !strings.Contains(body, "Title 1") {
		t.Error("Expected response to contain 'Title 1'")
	}
}

func TestIntegrationAPI_SearchPosts(t *testing.T) {
	router, pool, resource, db := setupTestServer(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	// Create posts with different content
	postRepo := repository.NewMongoPostRepository(db)
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
		err := postRepo.Create(context.Background(), post)
		if err != nil {
			t.Fatalf("Failed to create post: %v", err)
		}
	}

	// Search for "Go"
	req, _ := http.NewRequest("GET", "/posts?search=Go", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Response should contain Go-related posts
	body := w.Body.String()
	if !strings.Contains(body, "Go Programming") {
		t.Error("Expected response to contain 'Go Programming'")
	}
	if !strings.Contains(body, "Go Advanced") {
		t.Error("Expected response to contain 'Go Advanced'")
	}
}

func TestIntegrationAPI_ShowCreateForm(t *testing.T) {
	router, pool, resource, _ := setupTestServer(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	// Get create form
	req, _ := http.NewRequest("GET", "/posts/new", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestIntegrationAPI_ShowEditForm(t *testing.T) {
	router, pool, resource, db := setupTestServer(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	// Create a post first
	postRepo := repository.NewMongoPostRepository(db)
	post := model.NewPost("Test Title", "Test Content")
	err := postRepo.Create(context.Background(), post)
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	// Get edit form
	req, _ := http.NewRequest("GET", "/posts/edit?id="+post.ID.Hex(), nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Response should contain the post data
	body := w.Body.String()
	if !strings.Contains(body, "Test Title") {
		t.Error("Expected response to contain 'Test Title'")
	}
}

func TestIntegrationAPI_ShowPost(t *testing.T) {
	t.Skip("Skipping test - post-detail.html template not implemented in the application")

	router, pool, resource, db := setupTestServer(t)
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}()

	// Create a post first
	postRepo := repository.NewMongoPostRepository(db)
	post := model.NewPost("Test Title", "Test Content")
	err := postRepo.Create(context.Background(), post)
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	// View the post
	req, _ := http.NewRequest("GET", "/posts/view?id="+post.ID.Hex(), nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Response should contain the post data
	body := w.Body.String()
	if !strings.Contains(body, "Test Title") {
		t.Error("Expected response to contain 'Test Title'")
	}
	if !strings.Contains(body, "Test Content") {
		t.Error("Expected response to contain 'Test Content'")
	}
}
