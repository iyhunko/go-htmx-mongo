package handler

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/iyhunko/go-htmx-mongo/internal/domain"
	"github.com/iyhunko/go-htmx-mongo/internal/service"
)

// PostHandler handles HTTP requests for posts
type PostHandler struct {
	service   *service.PostService
	templates *template.Template
}

// NewPostHandler creates a new post handler
func NewPostHandler(service *service.PostService, templates *template.Template) *PostHandler {
	return &PostHandler{
		service:   service,
		templates: templates,
	}
}

// Index shows the home page with all posts
func (h *PostHandler) Index(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	search := r.URL.Query().Get("search")
	pageSize := 10

	var posts []*domain.Post
	var totalPages int
	var err error

	if search != "" {
		posts, totalPages, err = h.service.SearchPosts(r.Context(), search, page, pageSize)
	} else {
		posts, totalPages, err = h.service.GetPosts(r.Context(), page, pageSize)
	}

	if err != nil {
		log.Printf("Error getting posts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Posts":       posts,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"Search":      search,
	}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// PostsList returns the posts list partial for HTMX
func (h *PostHandler) PostsList(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	search := r.URL.Query().Get("search")
	pageSize := 10

	var posts []*domain.Post
	var totalPages int
	var err error

	if search != "" {
		posts, totalPages, err = h.service.SearchPosts(r.Context(), search, page, pageSize)
	} else {
		posts, totalPages, err = h.service.GetPosts(r.Context(), page, pageSize)
	}

	if err != nil {
		log.Printf("Error getting posts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Posts":       posts,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"Search":      search,
	}

	if err := h.templates.ExecuteTemplate(w, "posts-list.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ShowCreateForm shows the create post form
func (h *PostHandler) ShowCreateForm(w http.ResponseWriter, r *http.Request) {
	if err := h.templates.ExecuteTemplate(w, "post-form.html", map[string]interface{}{"Mode": "create"}); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// CreatePost handles post creation
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	post, err := h.service.CreatePost(r.Context(), title, content)
	if err != nil {
		data := map[string]interface{}{
			"Mode":    "create",
			"Error":   err.Error(),
			"Title":   title,
			"Content": content,
		}
		w.WriteHeader(http.StatusBadRequest)
		if err := h.templates.ExecuteTemplate(w, "post-form.html", data); err != nil {
			log.Printf("Error executing template: %v", err)
		}
		return
	}

	// Return the new post row
	data := map[string]interface{}{
		"Post": post,
	}
	w.Header().Set("HX-Trigger", "postCreated")
	if err := h.templates.ExecuteTemplate(w, "post-row.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ShowPost displays a single post
func (h *PostHandler) ShowPost(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
		return
	}

	post, err := h.service.GetPost(r.Context(), id)
	if err != nil {
		log.Printf("Error getting post: %v", err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Post": post,
	}

	if err := h.templates.ExecuteTemplate(w, "post-detail.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ShowEditForm shows the edit post form
func (h *PostHandler) ShowEditForm(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
		return
	}

	post, err := h.service.GetPost(r.Context(), id)
	if err != nil {
		log.Printf("Error getting post: %v", err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Mode": "edit",
		"Post": post,
	}

	if err := h.templates.ExecuteTemplate(w, "post-form.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// UpdatePost handles post update
func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	title := r.FormValue("title")
	content := r.FormValue("content")

	post, err := h.service.UpdatePost(r.Context(), id, title, content)
	if err != nil {
		// Get the original post to display in form
		originalPost, _ := h.service.GetPost(r.Context(), id)
		if originalPost == nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		data := map[string]interface{}{
			"Mode":    "edit",
			"Post":    originalPost,
			"Error":   err.Error(),
			"Title":   title,
			"Content": content,
		}
		w.WriteHeader(http.StatusBadRequest)
		if err := h.templates.ExecuteTemplate(w, "post-form.html", data); err != nil {
			log.Printf("Error executing template: %v", err)
		}
		return
	}

	// Return the updated post row
	data := map[string]interface{}{
		"Post": post,
	}
	if err := h.templates.ExecuteTemplate(w, "post-row.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// DeletePost handles post deletion
func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeletePost(r.Context(), id); err != nil {
		log.Printf("Error deleting post: %v", err)
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	// Return empty response - HTMX will remove the row
	w.WriteHeader(http.StatusOK)
}
