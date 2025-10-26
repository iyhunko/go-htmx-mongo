package domain

import (
	"strings"
	"testing"
)

func TestPostValidate(t *testing.T) {
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
			name:        "whitespace only title",
			title:       "   ",
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
			name:        "whitespace only content",
			title:       "Test Title",
			content:     "   ",
			wantErr:     true,
			expectedErr: "content is required",
		},
		{
			name:        "title too long",
			title:       strings.Repeat("a", 201),
			content:     "Test Content",
			wantErr:     true,
			expectedErr: "title must be less than 200 characters",
		},
		{
			name:        "content too long",
			title:       "Test Title",
			content:     strings.Repeat("a", 10001),
			wantErr:     true,
			expectedErr: "content must be less than 10000 characters",
		},
		{
			name:    "title at max length",
			title:   strings.Repeat("a", 200),
			content: "Test Content",
			wantErr: false,
		},
		{
			name:    "content at max length",
			title:   "Test Title",
			content: strings.Repeat("a", 10000),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post := NewPost(tt.title, tt.content)
			err := post.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
				} else if tt.expectedErr != "" && err.Error() != tt.expectedErr {
					t.Errorf("Validate() error = %v, want %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestNewPost(t *testing.T) {
	title := "  Test Title  "
	content := "  Test Content  "

	post := NewPost(title, content)

	if post.Title != "Test Title" {
		t.Errorf("NewPost() title = %v, want Test Title", post.Title)
	}

	if post.Content != "Test Content" {
		t.Errorf("NewPost() content = %v, want Test Content", post.Content)
	}

	if post.CreatedAt.IsZero() {
		t.Errorf("NewPost() CreatedAt should not be zero")
	}

	if post.UpdatedAt.IsZero() {
		t.Errorf("NewPost() UpdatedAt should not be zero")
	}

	if post.CreatedAt != post.UpdatedAt {
		t.Errorf("NewPost() CreatedAt should equal UpdatedAt")
	}
}

func TestPostUpdate(t *testing.T) {
	post := NewPost("Original Title", "Original Content")
	originalCreatedAt := post.CreatedAt
	originalUpdatedAt := post.UpdatedAt

	newTitle := "  Updated Title  "
	newContent := "  Updated Content  "

	post.Update(newTitle, newContent)

	if post.Title != "Updated Title" {
		t.Errorf("Update() title = %v, want Updated Title", post.Title)
	}

	if post.Content != "Updated Content" {
		t.Errorf("Update() content = %v, want Updated Content", post.Content)
	}

	if post.CreatedAt != originalCreatedAt {
		t.Errorf("Update() should not change CreatedAt")
	}

	if post.UpdatedAt == originalUpdatedAt {
		t.Errorf("Update() should change UpdatedAt")
	}

	if post.UpdatedAt.Before(originalUpdatedAt) {
		t.Errorf("Update() UpdatedAt should be after original UpdatedAt")
	}
}
