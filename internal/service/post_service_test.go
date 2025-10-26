package service

import (
	"context"
	"errors"
	"testing"

	"github.com/iyhunko/go-htmx-mongo/internal/model"
)

// mockPostRepository is a mock implementation for testing
type mockPostRepository struct {
	createFunc      func(ctx context.Context, post *model.Post) error
	findByIDFunc    func(ctx context.Context, id string) (*model.Post, error)
	findAllFunc     func(ctx context.Context, limit, offset int) ([]*model.Post, error)
	searchFunc      func(ctx context.Context, query string, limit, offset int) ([]*model.Post, error)
	updateFunc      func(ctx context.Context, post *model.Post) error
	deleteFunc      func(ctx context.Context, id string) error
	countFunc       func(ctx context.Context) (int64, error)
	countSearchFunc func(ctx context.Context, query string) (int64, error)
}

func (m *mockPostRepository) Create(ctx context.Context, post *model.Post) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, post)
	}
	return nil
}

func (m *mockPostRepository) FindByID(ctx context.Context, id string) (*model.Post, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockPostRepository) FindAll(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, limit, offset)
	}
	return nil, nil
}

func (m *mockPostRepository) Search(ctx context.Context, query string, limit, offset int) ([]*model.Post, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, query, limit, offset)
	}
	return nil, nil
}

func (m *mockPostRepository) Update(ctx context.Context, post *model.Post) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, post)
	}
	return nil
}

func (m *mockPostRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockPostRepository) Count(ctx context.Context) (int64, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx)
	}
	return 0, nil
}

func (m *mockPostRepository) CountSearch(ctx context.Context, query string) (int64, error) {
	if m.countSearchFunc != nil {
		return m.countSearchFunc(ctx, query)
	}
	return 0, nil
}

func TestCreatePost(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		content     string
		wantErr     bool
		expectedErr string
	}{
		{
			name:    "valid post",
			title:   "Test Title",
			content: "Test Content",
			wantErr: false,
		},
		{
			name:        "empty title",
			title:       "",
			content:     "Test Content",
			wantErr:     true,
			expectedErr: "title is required",
		},
		{
			name:        "empty content",
			title:       "Test Title",
			content:     "",
			wantErr:     true,
			expectedErr: "content is required",
		},
		{
			name:        "title too long",
			title:       string(make([]byte, 201)),
			content:     "Test Content",
			wantErr:     true,
			expectedErr: "title must be less than 200 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPostRepository{}
			service := NewPostService(repo)

			post, err := service.CreatePost(context.Background(), tt.title, tt.content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreatePost() expected error but got nil")
				} else if tt.expectedErr != "" && err.Error() != tt.expectedErr {
					t.Errorf("CreatePost() error = %v, want %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("CreatePost() unexpected error = %v", err)
				}
				if post == nil {
					t.Errorf("CreatePost() expected post but got nil")
				}
			}
		})
	}
}

func TestGetPosts(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		expectedPage int
		expectedSize int
	}{
		{
			name:         "valid pagination",
			page:         1,
			pageSize:     10,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name:         "invalid page defaults to 1",
			page:         0,
			pageSize:     10,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name:         "invalid page size defaults to 10",
			page:         1,
			pageSize:     0,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name:         "page size over limit capped at 100",
			page:         1,
			pageSize:     150,
			expectedPage: 1,
			expectedSize: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPostRepository{
				findAllFunc: func(ctx context.Context, limit, offset int) ([]*model.Post, error) {
					if limit != tt.expectedSize {
						t.Errorf("FindAll() limit = %v, want %v", limit, tt.expectedSize)
					}
					expectedOffset := (tt.expectedPage - 1) * tt.expectedSize
					if offset != expectedOffset {
						t.Errorf("FindAll() offset = %v, want %v", offset, expectedOffset)
					}
					return []*model.Post{}, nil
				},
				countFunc: func(ctx context.Context) (int64, error) {
					return 0, nil
				},
			}
			service := NewPostService(repo)

			_, _, err := service.GetPosts(context.Background(), tt.page, tt.pageSize)
			if err != nil {
				t.Errorf("GetPosts() unexpected error = %v", err)
			}
		})
	}
}

func TestUpdatePost(t *testing.T) {
	existingPost := model.NewPost("Original Title", "Original Content")

	tests := []struct {
		name        string
		id          string
		title       string
		content     string
		wantErr     bool
		expectedErr string
	}{
		{
			name:    "valid update",
			id:      "123",
			title:   "Updated Title",
			content: "Updated Content",
			wantErr: false,
		},
		{
			name:        "empty title",
			id:          "123",
			title:       "",
			content:     "Updated Content",
			wantErr:     true,
			expectedErr: "title is required",
		},
		{
			name:        "empty content",
			id:          "123",
			title:       "Updated Title",
			content:     "",
			wantErr:     true,
			expectedErr: "content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPostRepository{
				findByIDFunc: func(ctx context.Context, id string) (*model.Post, error) {
					return existingPost, nil
				},
			}
			service := NewPostService(repo)

			_, err := service.UpdatePost(context.Background(), tt.id, tt.title, tt.content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdatePost() expected error but got nil")
				} else if tt.expectedErr != "" && err.Error() != tt.expectedErr {
					t.Errorf("UpdatePost() error = %v, want %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("UpdatePost() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestSearchPosts(t *testing.T) {
	repo := &mockPostRepository{
		searchFunc: func(ctx context.Context, query string, limit, offset int) ([]*model.Post, error) {
			if query != "test" {
				t.Errorf("Search() query = %v, want test", query)
			}
			return []*model.Post{}, nil
		},
		countSearchFunc: func(ctx context.Context, query string) (int64, error) {
			return 0, nil
		},
	}
	service := NewPostService(repo)

	_, _, err := service.SearchPosts(context.Background(), "test", 1, 10)
	if err != nil {
		t.Errorf("SearchPosts() unexpected error = %v", err)
	}
}
