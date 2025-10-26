package http

import (
	"github.com/gin-gonic/gin"
	"github.com/iyhunko/go-htmx-mongo/internal/controller"
)

// SetupRoutes configures all application routes
func SetupRoutes(router *gin.Engine, postController *controller.PostController) {
	// Serve static files
	router.Static("/static", "web/static")

	// Post routes
	router.GET("/", postController.Index)
	router.GET("/posts", postController.PostsList)
	router.POST("/posts", postController.CreatePost)
	router.PUT("/posts", postController.UpdatePost)
	router.DELETE("/posts/:id", postController.DeletePost)
	router.GET("/posts/new", postController.ShowCreateForm)
	router.GET("/posts/edit", postController.ShowEditForm)
	router.GET("/posts/view", postController.ShowPost)
}
