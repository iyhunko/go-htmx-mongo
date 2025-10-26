package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	MongoURI   string
	MongoDB    string
	ServerPort string
}

// Load loads configuration from environment variables
// It returns a Config struct populated with values from environment variables
// or default values if environment variables are not set.
func Load() *Config {
	return &Config{
		MongoURI:   getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:    getEnv("MONGO_DB", "newsdb"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
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
