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

// CreatePost creates a new post
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

// GetPost retrieves a post by ID
func (s *PostService) GetPost(ctx context.Context, id string) (*domain.Post, error) {
	post, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return post, nil
}

// GetPosts retrieves paginated posts
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

// SearchPosts searches posts by query
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

// UpdatePost updates an existing post
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

// DeletePost deletes a post
func (s *PostService) DeletePost(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
