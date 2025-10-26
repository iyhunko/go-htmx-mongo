package controller

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iyhunko/go-htmx-mongo/internal/domain"
	"github.com/iyhunko/go-htmx-mongo/internal/service"
	"github.com/iyhunko/go-htmx-mongo/pkg/config"
)

// PostController handles HTTP requests for posts
type PostController struct {
	service   *service.PostService
	templates *template.Template
	config    *config.Config
}

// NewPostController creates a new post controller
func NewPostController(service *service.PostService, templates *template.Template, cfg *config.Config) *PostController {
	return &PostController{
		service:   service,
		templates: templates,
		config:    cfg,
	}
}

// Index shows the home page with all posts
func (c *PostController) Index(ctx *gin.Context) {
	page := 1
	if p := ctx.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	search := ctx.Query("search")
	pageSize := c.config.PageSizeLimit

	var posts []*domain.Post
	var totalPages int
	var err error

	if search != "" {
		posts, totalPages, err = c.service.SearchPosts(ctx.Request.Context(), search, page, pageSize)
	} else {
		posts, totalPages, err = c.service.GetPosts(ctx.Request.Context(), page, pageSize)
	}

	if err != nil {
		slog.Error("Failed to get posts", "error", err, "page", page, "search", search)
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": "Internal server error"})
		return
	}

	data := map[string]interface{}{
		"Posts":       posts,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"Search":      search,
	}

	if err := c.templates.ExecuteTemplate(ctx.Writer, "index.html", data); err != nil {
		slog.Error("Failed to execute template", "error", err, "template", "index.html")
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": "Internal server error"})
	}
}

// PostsList returns the posts list partial for HTMX
func (c *PostController) PostsList(ctx *gin.Context) {
	page := 1
	if p := ctx.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	search := ctx.Query("search")
	pageSize := c.config.PageSizeLimit

	var posts []*domain.Post
	var totalPages int
	var err error

	if search != "" {
		posts, totalPages, err = c.service.SearchPosts(ctx.Request.Context(), search, page, pageSize)
	} else {
		posts, totalPages, err = c.service.GetPosts(ctx.Request.Context(), page, pageSize)
	}

	if err != nil {
		slog.Error("Failed to get posts", "error", err, "page", page, "search", search)
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": "Internal server error"})
		return
	}

	data := map[string]interface{}{
		"Posts":       posts,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"Search":      search,
	}

	if err := c.templates.ExecuteTemplate(ctx.Writer, "posts-list.html", data); err != nil {
		slog.Error("Failed to execute template", "error", err, "template", "posts-list.html")
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": "Internal server error"})
	}
}

// ShowCreateForm shows the create post form
func (c *PostController) ShowCreateForm(ctx *gin.Context) {
	if err := c.templates.ExecuteTemplate(ctx.Writer, "post-form.html", map[string]interface{}{"Mode": "create"}); err != nil {
		slog.Error("Failed to execute template", "error", err, "template", "post-form.html")
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": "Internal server error"})
	}
}

// CreatePost handles post creation
func (c *PostController) CreatePost(ctx *gin.Context) {
	title := ctx.PostForm("title")
	content := ctx.PostForm("content")

	slog.Info("Creating new post", "title", title)

	post, err := c.service.CreatePost(ctx.Request.Context(), title, content)
	if err != nil {
		slog.Warn("Failed to create post", "error", err, "title", title)
		data := map[string]interface{}{
			"Mode":    "create",
			"Error":   err.Error(),
			"Title":   title,
			"Content": content,
		}
		ctx.Writer.WriteHeader(http.StatusBadRequest)
		if err := c.templates.ExecuteTemplate(ctx.Writer, "post-form.html", data); err != nil {
			slog.Error("Failed to execute template", "error", err, "template", "post-form.html")
		}
		return
	}

	slog.Info("Post created successfully", "id", post.ID.Hex(), "title", post.Title)

	// Return the new post row
	data := map[string]interface{}{
		"Post": post,
	}
	ctx.Writer.Header().Set("HX-Trigger", "postCreated")
	if err := c.templates.ExecuteTemplate(ctx.Writer, "post-row.html", data); err != nil {
		slog.Error("Failed to execute template", "error", err, "template", "post-row.html")
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": "Internal server error"})
	}
}

// ShowPost displays a single post
func (c *PostController) ShowPost(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		slog.Warn("Post ID required but not provided")
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"Error": "Post ID required"})
		return
	}

	post, err := c.service.GetPost(ctx.Request.Context(), id)
	if err != nil {
		slog.Warn("Post not found", "id", id, "error", err)
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{"Error": "Post not found"})
		return
	}

	data := map[string]interface{}{
		"Post": post,
	}

	if err := c.templates.ExecuteTemplate(ctx.Writer, "post-detail.html", data); err != nil {
		slog.Error("Failed to execute template", "error", err, "template", "post-detail.html")
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": "Internal server error"})
	}
}

// ShowEditForm shows the edit post form
func (c *PostController) ShowEditForm(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		slog.Warn("Post ID required but not provided")
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"Error": "Post ID required"})
		return
	}

	post, err := c.service.GetPost(ctx.Request.Context(), id)
	if err != nil {
		slog.Warn("Post not found for edit", "id", id, "error", err)
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{"Error": "Post not found"})
		return
	}

	data := map[string]interface{}{
		"Mode": "edit",
		"Post": post,
	}

	if err := c.templates.ExecuteTemplate(ctx.Writer, "post-form.html", data); err != nil {
		slog.Error("Failed to execute template", "error", err, "template", "post-form.html")
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": "Internal server error"})
	}
}

// UpdatePost handles post update
func (c *PostController) UpdatePost(ctx *gin.Context) {
	id := ctx.Param("id")
	title := ctx.PostForm("title")
	content := ctx.PostForm("content")

	slog.Info("Updating post", "id", id, "title", title)

	post, err := c.service.UpdatePost(ctx.Request.Context(), id, title, content)
	if err != nil {
		slog.Warn("Failed to update post", "id", id, "error", err)
		// Get the original post to display in form
		originalPost, _ := c.service.GetPost(ctx.Request.Context(), id)
		if originalPost == nil {
			ctx.HTML(http.StatusNotFound, "error.html", gin.H{"Error": "Post not found"})
			return
		}

		data := map[string]interface{}{
			"Mode":    "edit",
			"Post":    originalPost,
			"Error":   err.Error(),
			"Title":   title,
			"Content": content,
		}
		ctx.Writer.WriteHeader(http.StatusBadRequest)
		if err := c.templates.ExecuteTemplate(ctx.Writer, "post-form.html", data); err != nil {
			slog.Error("Failed to execute template", "error", err, "template", "post-form.html")
		}
		return
	}

	slog.Info("Post updated successfully", "id", post.ID.Hex(), "title", post.Title)

	// Return the updated post row
	data := map[string]interface{}{
		"Post": post,
	}
	if err := c.templates.ExecuteTemplate(ctx.Writer, "post-row.html", data); err != nil {
		slog.Error("Failed to execute template", "error", err, "template", "post-row.html")
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": "Internal server error"})
	}
}

// DeletePost handles post deletion
func (c *PostController) DeletePost(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		slog.Warn("Post ID required but not provided for deletion")
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"Error": "Post ID required"})
		return
	}

	slog.Info("Deleting post", "id", id)

	if err := c.service.DeletePost(ctx.Request.Context(), id); err != nil {
		slog.Error("Failed to delete post", "id", id, "error", err)
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": "Failed to delete post"})
		return
	}

	slog.Info("Post deleted successfully", "id", id)

	// Return empty response - HTMX will remove the row
	ctx.Status(http.StatusOK)
}
