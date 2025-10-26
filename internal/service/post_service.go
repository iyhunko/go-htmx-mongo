package service

import (
	"context"
	"errors"

	"github.com/iyhunko/go-htmx-mongo/internal/domain"
)

var (
	ErrPostNotFound     = errors.New("post not found")
	ErrInvalidID        = errors.New("invalid post id")
	ErrValidationFailed = errors.New("validation failed")
)

// PostService handles business logic for posts
type PostService struct {
	repo domain.PostRepository
}

// NewPostService creates a new post service
func NewPostService(repo domain.PostRepository) *PostService {
	return &PostService{
		repo: repo,
	}
}

// CreatePost creates a new post with the provided title and content.
// It validates the post before saving it to the repository.
// Returns the created post or an error if validation or creation fails.
func (s *PostService) CreatePost(ctx context.Context, title, content string) (*domain.Post, error) {
	post := domain.NewPost(title, content)

	if err := post.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

// GetPost retrieves a post by its ID.
// Returns the post if found, or an error if the post doesn't exist or the ID is invalid.
func (s *PostService) GetPost(ctx context.Context, id string) (*domain.Post, error) {
	post, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return post, nil
}

// GetPosts retrieves paginated posts.
// It returns posts for the requested page, the total number of pages, and an error if any.
// Page numbers start at 1, and invalid values are adjusted to defaults (page=1, pageSize=10).
// Maximum page size is capped at 100.
func (s *PostService) GetPosts(ctx context.Context, page, pageSize int) ([]*domain.Post, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	posts, err := s.repo.FindAll(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return posts, totalPages, nil
}

// SearchPosts searches posts by query string in title and content.
// It returns matching posts for the requested page, the total number of pages, and an error if any.
// Page numbers start at 1, and invalid values are adjusted to defaults (page=1, pageSize=10).
// Maximum page size is capped at 100.
func (s *PostService) SearchPosts(ctx context.Context, query string, page, pageSize int) ([]*domain.Post, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	posts, err := s.repo.Search(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountSearch(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return posts, totalPages, nil
}

// UpdatePost updates an existing post with new title and content.
// It retrieves the post, updates it, validates it, and saves it back to the repository.
// Returns the updated post or an error if the post is not found or validation fails.
func (s *PostService) UpdatePost(ctx context.Context, id, title, content string) (*domain.Post, error) {
	post, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	post.Update(title, content)

	if err := post.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

// DeletePost deletes a post by its ID.
// Returns an error if the post is not found or deletion fails.
func (s *PostService) DeletePost(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
