# News Articles Service - Go + HTMX + MongoDB

A modern web application for managing news articles, built with Go, HTMX for dynamic interactions, and MongoDB for data persistence.

## Features

- **CRUD Operations**: Create, Read, Update, and Delete news articles
- **Pagination**: Efficiently browse through large sets of articles
- **Search**: Find articles by title or content
- **Server-Side Rendering**: Fast initial page loads with HTMX for dynamic updates
- **Data Validation**: Ensure data integrity with built-in validation (validation happens on form submission)
- **Responsive UI**: Clean, modern interface powered by HTMX
- **Auto-Migration**: Automatic MongoDB collection and index creation on startup
- **Structured Logging**: JSON-formatted structured logging using Go's standard `slog` package
- **Dockerized**: Easy deployment with Docker Compose

## Tech Stack

- **Backend**: Go (Golang) with Gin framework
- **Frontend**: HTML + HTMX for dynamic interactions
- **Database**: MongoDB with automatic migration
- **Logging**: Structured logging with `slog`
- **Testing**: Go testing framework + Dockertest for integration tests
- **Containerization**: Docker and Docker Compose

## Project Structure

```
.
├── cmd/
│   └── server/          # Application entry point
├── integration/
│   ├── helper.go        # Integration test helper functions
│   ├── api_integration_test.go      # API integration tests
│   └── repository_integration_test.go  # Repository integration tests
├── http/
│   └── routes.go        # HTTP route definitions
├── internal/
│   ├── controller/      # HTTP request controllers (formerly handlers)
│   ├── db/              # Database connection and migration
│   ├── domain/          # Domain models and interfaces
│   ├── repository/      # Data access layer (MongoDB)
│   └── service/         # Business logic layer
├── pkg/
│   └── config/          # Configuration management
├── web/
│   ├── templates/       # HTML templates
│   ├── static/          # Static assets
│   └── templates.go     # Template loading with custom functions
├── docker-compose.yaml  # Docker Compose configuration
├── Dockerfile           # Application container definition
├── Makefile             # Build and test commands
└── go.mod               # Go module dependencies
```

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose (for running MongoDB and full stack)
- Make (optional, for using Makefile commands)

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/iyhunko/go-htmx-mongo.git
cd go-htmx-mongo
```

### 2. Start with Docker Compose (Recommended)

Using Docker Compose to run both MongoDB and the application:
```bash
make docker-up
# Or manually:
docker-compose up -d
```

The application will be available at http://localhost:8080

### 3. Or run locally with MongoDB in Docker

Start MongoDB only:
```bash
docker run -d --name mongo-news -p 27017:27017 mongo:7
```

Run the application:
```bash
make run
# Or:
go run cmd/server/main.go
```

### 4. Stop services

```bash
make docker-down
# Or manually:
docker-compose down
```

## Configuration

The application uses environment variables for configuration:

- `MONGO_URI`: MongoDB connection string (default: `mongodb://localhost:27017`)
- `MONGO_DB`: Database name (default: `newsdb`)
- `SERVER_PORT`: HTTP server port (default: `8080`)

Example:
```bash
export MONGO_URI=mongodb://localhost:27017
export MONGO_DB=newsdb
export SERVER_PORT=8080
go run cmd/server/main.go
```

## Development

### Available Make Commands

```bash
make help              # Display available commands
make build             # Build the application
make run               # Run the application
make test              # Run all tests
make test-unit         # Run unit tests only
make test-integration  # Run integration tests
make coverage          # Generate test coverage report
make lint              # Run linter
make fmt               # Format code
make vet               # Run go vet
make tidy              # Tidy go modules
make docker-up         # Start services with docker-compose
make docker-down       # Stop docker-compose services
make clean             # Clean build artifacts
```

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests (requires Docker)
make test-integration

# Generate coverage report
make coverage
```

### Building

```bash
# Build the application
make build

# The binary will be available at ./bin/server
./bin/server
```

## API Endpoints

The application uses HTMX for server-side rendering. All endpoints return HTML:

- `GET /` - Home page with posts list
- `GET /posts` - Get posts list (with pagination and search)
- `GET /posts/new` - Show create post form
- `POST /posts` - Create a new post
- `GET /posts/edit?id={id}` - Show edit post form
- `PUT /posts` - Update a post
- `DELETE /posts?id={id}` - Delete a post

### Query Parameters

- `page`: Page number for pagination (default: 1)
- `search`: Search query for filtering posts
- `id`: Post ID for single post operations

## Data Model

### Post

```go
type Post struct {
    ID        primitive.ObjectID
    Title     string
    Content   string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Validation Rules

- **Title**: Required, 1-200 characters
- **Content**: Required, 1-10,000 characters
- **Validation**: Server-side validation occurs on form submission (not on field change)

## Database Auto-Migration

The application automatically creates MongoDB collections and indexes on startup:

- **Collections**: `posts` collection is created if it doesn't exist
- **Indexes**: 
  - Index on `created_at` field for sorting
  - Text index on `title` and `content` fields for search functionality

## Logging

The application uses Go's standard `slog` package for structured logging with JSON output:

- Request/response logging with method, path, status, and duration
- Database connection and migration events
- Error logging with context

## Testing

The project includes comprehensive test coverage:

- **Unit Tests**: Test business logic in isolation
- **Integration Tests**: Test database operations with real MongoDB using Dockertest

Integration tests automatically:
- Start a MongoDB container
- Run tests against the real database
- Clean up the container after tests

## Docker Deployment

### Build and run with Docker Compose

```bash
docker-compose up -d
```

This will:
1. Start a MongoDB container
2. Build and start the Go application container
3. Configure networking between containers
4. Expose the application on port 8080

### Access the application

Open http://localhost:8080 in your browser.

### View logs

```bash
docker-compose logs -f app
```

## License

MIT