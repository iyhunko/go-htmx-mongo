package repository

import (
	"context"

	"github.com/iyhunko/go-htmx-mongo/internal/model"
)

// PostRepository defines the interface for post data operations
type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	FindByID(ctx context.Context, id string) (*model.Post, error)
	FindAll(ctx context.Context, limit, offset int) ([]*model.Post, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*model.Post, error)
	Update(ctx context.Context, post *model.Post) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
	CountSearch(ctx context.Context, query string) (int64, error)
}
