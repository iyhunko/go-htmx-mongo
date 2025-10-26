package model

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post represents a news article
type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// Validate validates post fields.
// It checks that title and content are not empty and within length limits.
// Returns an error describing the validation failure, or nil if validation passes.
func (p *Post) Validate() error {
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("title is required")
	}
	if len(p.Title) > 200 {
		return errors.New("title must be less than 200 characters")
	}
	if strings.TrimSpace(p.Content) == "" {
		return errors.New("content is required")
	}
	if len(p.Content) > 10000 {
		return errors.New("content must be less than 10000 characters")
	}
	return nil
}

// NewPost creates a new post with timestamps.
// It trims whitespace from title and content and sets CreatedAt and UpdatedAt to current time.
func NewPost(title, content string) *Post {
	now := time.Now()
	return &Post{
		Title:     strings.TrimSpace(title),
		Content:   strings.TrimSpace(content),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update updates the post content and timestamp.
// It trims whitespace from title and content and updates UpdatedAt to current time.
func (p *Post) Update(title, content string) {
	p.Title = strings.TrimSpace(title)
	p.Content = strings.TrimSpace(content)
	p.UpdatedAt = time.Now()
}
