package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB holds the database connection and client
type MongoDB struct {
	Client *mongo.Client
	DB     *mongo.Database
}

// Connect establishes a connection to MongoDB and performs auto-migration
func Connect(ctx context.Context, uri, dbName string) (*MongoDB, error) {
	slog.Info("Connecting to MongoDB", "uri", uri, "database", dbName)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping MongoDB to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	slog.Info("Successfully connected to MongoDB")

	db := client.Database(dbName)

	// Perform auto-migration
	if err := migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &MongoDB{
		Client: client,
		DB:     db,
	}, nil
}

// Disconnect closes the MongoDB connection
func (m *MongoDB) Disconnect(ctx context.Context) error {
	if m.Client != nil {
		slog.Info("Disconnecting from MongoDB")
		return m.Client.Disconnect(ctx)
	}
	return nil
}

// migrate performs auto-migration of MongoDB collections and indexes
func migrate(ctx context.Context, db *mongo.Database) error {
	slog.Info("Starting database migration")

	// Create posts collection if it doesn't exist
	if err := createCollection(ctx, db, "posts"); err != nil {
		return err
	}

	// Create indexes for posts collection
	if err := createPostsIndexes(ctx, db); err != nil {
		return err
	}

	slog.Info("Database migration completed successfully")
	return nil
}

// createCollection creates a collection if it doesn't exist
func createCollection(ctx context.Context, db *mongo.Database, name string) error {
	collections, err := db.ListCollectionNames(ctx, bson.M{"name": name})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	if len(collections) == 0 {
		slog.Info("Creating collection", "name", name)
		if err := db.CreateCollection(ctx, name); err != nil {
			return fmt.Errorf("failed to create collection %s: %w", name, err)
		}
		slog.Info("Collection created successfully", "name", name)
	} else {
		slog.Info("Collection already exists", "name", name)
	}

	return nil
}

// createPostsIndexes creates indexes for the posts collection
func createPostsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("posts")

	// Create index on created_at for sorting
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: -1}},
		Options: options.Index().
			SetName("created_at_desc"),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create created_at index: %w", err)
	}
	slog.Info("Created index on posts.created_at")

	// Create text index for search on title and content
	textIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "title", Value: "text"},
			{Key: "content", Value: "text"},
		},
		Options: options.Index().
			SetName("title_content_text"),
	}

	_, err = collection.Indexes().CreateOne(ctx, textIndexModel)
	if err != nil {
		return fmt.Errorf("failed to create text index: %w", err)
	}
	slog.Info("Created text index on posts.title and posts.content")

	return nil
}

// HealthCheck verifies the database connection is healthy
func (m *MongoDB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := m.Client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}
