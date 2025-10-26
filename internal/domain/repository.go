package domain

import "context"

// PostRepository defines the interface for post data operations
type PostRepository interface {
	Create(ctx context.Context, post *Post) error
	FindByID(ctx context.Context, id string) (*Post, error)
	FindAll(ctx context.Context, limit, offset int) ([]*Post, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*Post, error)
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
	CountSearch(ctx context.Context, query string) (int64, error)
}
