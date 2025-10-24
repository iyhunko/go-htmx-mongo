package repository

import (
	"context"
	"errors"
	"time"

	"github.com/iyhunko/go-htmx-mongo/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrPostNotFound = errors.New("post not found")
	ErrInvalidID    = errors.New("invalid post id")
)

type mongoPostRepository struct {
	collection *mongo.Collection
}

// NewMongoPostRepository creates a new MongoDB post repository
func NewMongoPostRepository(db *mongo.Database) domain.PostRepository {
	return &mongoPostRepository{
		collection: db.Collection("posts"),
	}
}

func (r *mongoPostRepository) Create(ctx context.Context, post *domain.Post) error {
	post.ID = primitive.NewObjectID()
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, post)
	return err
}

func (r *mongoPostRepository) FindByID(ctx context.Context, id string) (*domain.Post, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	var post domain.Post
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	return &post, nil
}

func (r *mongoPostRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Post, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []*domain.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *mongoPostRepository) Search(ctx context.Context, query string, limit, offset int) ([]*domain.Post, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"content": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []*domain.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *mongoPostRepository) Update(ctx context.Context, post *domain.Post) error {
	post.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"title":      post.Title,
			"content":    post.Content,
			"updated_at": post.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": post.ID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrPostNotFound
	}

	return nil
}

func (r *mongoPostRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidID
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrPostNotFound
	}

	return nil
}

func (r *mongoPostRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

func (r *mongoPostRepository) CountSearch(ctx context.Context, query string) (int64, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"content": bson.M{"$regex": query, "$options": "i"}},
		},
	}
	return r.collection.CountDocuments(ctx, filter)
}
