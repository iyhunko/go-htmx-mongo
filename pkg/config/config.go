package config

import (
	"net/url"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	HttpServerHost  string
	HttpServerPort  string
	MongoDBHost     string
	MongoDBPort     string
	MongoDBDatabase string
	MongoDBUser     string
	MongoDBPassword string
	PageSizeLimit   int
}

// Load loads configuration from environment variables
// It returns a Config struct populated with values from environment variables
// or default values if environment variables are not set.
func Load() *Config {
	return &Config{
		HttpServerHost:  getEnv("HTTP_SERVER_HOST", ""),
		HttpServerPort:  getEnv("HTTP_SERVER_PORT", "8080"),
		MongoDBHost:     getEnv("MONGODB_HOST", "localhost"),
		MongoDBPort:     getEnv("MONGODB_PORT", "27017"),
		MongoDBDatabase: getEnv("MONGODB_DATABASE", "newsdb"),
		MongoDBUser:     getEnv("MONGODB_USER", ""),
		MongoDBPassword: getEnv("MONGODB_PASSWORD", ""),
		PageSizeLimit:   getEnvAsInt("PAGE_SIZE_LIMIT", 100),
	}
}

// GetMongoURI constructs the MongoDB connection URI from config values
func (c *Config) GetMongoURI() string {
	var auth string
	if c.MongoDBUser != "" && c.MongoDBPassword != "" {
		auth = url.QueryEscape(c.MongoDBUser) + ":" + url.QueryEscape(c.MongoDBPassword) + "@"
	}
	return "mongodb://" + auth + c.MongoDBHost + ":" + c.MongoDBPort
}

// GetServerAddress constructs the HTTP server address from config values
func (c *Config) GetServerAddress() string {
	return c.HttpServerHost + ":" + c.HttpServerPort
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
