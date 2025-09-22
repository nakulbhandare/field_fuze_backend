# Development Guide

A comprehensive guide for developers working on the FieldFuze Backend project, covering setup, development workflow, testing, and contribution guidelines.

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Development Environment](#development-environment)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Code Quality](#code-quality)
- [Debugging](#debugging)
- [Contributing](#contributing)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites

- **Go**: 1.23.2 or higher
- **Git**: Latest version
- **AWS CLI**: For DynamoDB access (optional)
- **Docker**: For containerized development (optional)

### Clone and Setup

```bash
# Clone the repository
git clone <repository-url>
cd fieldfuze-backend

# Install dependencies
go mod download

# Copy configuration
cp config.json.example config.json

# Edit configuration with your settings
vim config.json

# Run the application
go run main.go
```

### Verify Installation

```bash
# Check if server is running
curl http://localhost:8081/api/v1/auth/health

# Expected response:
{
  "status": "healthy",
  "version": "1.0.0",
  "service": "FieldFuze Backend"
}
```

## Development Environment

### Configuration Files

#### config.json
Create your development configuration:

```json
{
  "app": {
    "name": "FieldFuze Backend",
    "version": "1.0.0",
    "env": "development",
    "host": "0.0.0.0",
    "port": "8081"
  },
  "jwt": {
    "secret": "your-development-secret-key",
    "expires_in": "30m"
  },
  "aws": {
    "region": "us-east-1",
    "access_key_id": "your-access-key",
    "secret_access_key": "your-secret-key",
    "dynamodb_endpoint": "http://localhost:8000",
    "dynamodb_table_prefix": "dev"
  },
  "logging": {
    "level": "debug",
    "format": "text"
  },
  "cors": {
    "origins": ["*"]
  },
  "basePath": "/api/v1/auth"
}
```

#### Environment Variables
Alternatively, use environment variables:

```bash
# Application
export APP_ENV=development
export APP_PORT=8081

# JWT
export JWT_SECRET=your-development-secret

# AWS (for local DynamoDB)
export AWS_REGION=us-east-1
export DYNAMODB_ENDPOINT=http://localhost:8000
export DYNAMODB_TABLE_PREFIX=dev

# Logging
export LOG_LEVEL=debug
export LOG_FORMAT=text
```

### Local DynamoDB Setup

#### Using Docker
```bash
# Start local DynamoDB
docker run -p 8000:8000 amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb

# Verify it's running
aws dynamodb list-tables --endpoint-url http://localhost:8000
```

#### Using AWS CLI
```bash
# Configure AWS CLI for local development
aws configure set aws_access_key_id dummy
aws configure set aws_secret_access_key dummy
aws configure set region us-east-1

# Create tables manually (optional, worker will create them)
aws dynamodb create-table \
  --table-name dev-users \
  --attribute-definitions \
    AttributeName=id,AttributeType=S \
  --key-schema \
    AttributeName=id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --endpoint-url http://localhost:8000
```

### IDE Setup

#### VS Code
Recommended extensions:

```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-json",
    "bradlc.vscode-tailwindcss",
    "formulahendry.auto-rename-tag",
    "ms-vscode.thunder-client"
  ]
}
```

#### Go Tools
Install essential Go tools:

```bash
# Install go tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest

# Update Swagger documentation
swag init
```

## Project Structure

```
fieldfuze-backend/
├── config.json                 # Configuration file
├── main.go                     # Application entry point
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
│
├── controller/                 # HTTP request handlers
│   ├── controller.go          # Main controller factory
│   ├── user_controller.go     # User management endpoints
│   ├── role_controller.go     # Role management endpoints
│   └── infrastructure_controller.go # Infrastructure endpoints
│
├── services/                   # Business logic layer
│   ├── services.go            # Service factory
│   ├── user_service.go        # User business logic
│   ├── role_service.go        # Role business logic
│   └── infrastructure_service.go # Infrastructure logic
│
├── repository/                 # Data access layer
│   ├── repository.go          # Repository factory
│   ├── user_repo.go           # User data access
│   └── role_repo.go           # Role data access
│
├── models/                     # Data structures
│   ├── user.go               # User models
│   ├── role.go               # Role models
│   ├── auth.go               # Authentication models
│   ├── config.go             # Configuration models
│   ├── worker.go             # Worker models
│   └── api.go                # API response models
│
├── middelware/                 # HTTP middleware
│   ├── auth.go               # Authentication & authorization
│   ├── cors.go               # CORS handling
│   └── logging.go            # Request logging
│
├── worker/                     # Background workers
│   ├── worker.go             # Worker management
│   ├── infrastructure.go     # Infrastructure setup
│   ├── status.go             # Status management
│   └── lock.go               # Distributed locking
│
├── utils/                      # Utility functions
│   ├── utils.go              # General utilities
│   ├── logger/               # Logging utilities
│   └── swagger/              # Swagger utilities
│
├── dal/                        # Data access layer
│   └── dal.go                # DynamoDB client
│
├── infrastructure/             # Infrastructure configuration
│   ├── roles.json            # Role definitions
│   ├── table_schema.json     # Table schemas
│   └── tables.go             # Table management
│
├── docs/                       # Documentation
│   ├── swagger.json          # Swagger specification
│   ├── swagger.yaml          # Swagger YAML
│   └── docs.go               # Generated docs
│
└── docs/                       # Project documentation
    ├── CONTROLLERS.md
    ├── SERVICES.md
    ├── MODELS.md
    └── ...
```

## Development Workflow

### Git Workflow

#### Branching Strategy
```bash
# Feature development
git checkout -b feature/your-feature-name
git commit -m "feat: add new feature"
git push origin feature/your-feature-name

# Bug fixes
git checkout -b fix/bug-description
git commit -m "fix: resolve bug issue"
git push origin fix/bug-description

# Documentation
git checkout -b docs/update-readme
git commit -m "docs: update README with setup instructions"
git push origin docs/update-readme
```

#### Commit Message Convention
Follow conventional commits:

```bash
# Feature
git commit -m "feat: add user role assignment API"

# Bug fix
git commit -m "fix: resolve JWT token expiration issue"

# Documentation
git commit -m "docs: update API documentation"

# Refactor
git commit -m "refactor: improve error handling in controllers"

# Test
git commit -m "test: add unit tests for user service"

# Build/CI
git commit -m "ci: update deployment pipeline"
```

### Development Commands

#### Common Tasks
```bash
# Run development server
go run main.go

# Run with hot reload (using air)
air

# Format code
go fmt ./...

# Update imports
goimports -w .

# Lint code
golangci-lint run

# Run tests
go test ./...

# Run tests with coverage
go test -v -cover ./...

# Update Swagger docs
swag init

# Build binary
go build -o fieldfuze-backend main.go
```

#### Makefile (Create this for convenience)
```makefile
.PHONY: run build test lint fmt swagger clean

# Development
run:
	go run main.go

dev:
	air

# Code quality
fmt:
	go fmt ./...
	goimports -w .

lint:
	golangci-lint run

# Testing
test:
	go test -v ./...

test-coverage:
	go test -v -cover ./...

# Documentation
swagger:
	swag init

# Build
build:
	go build -o bin/fieldfuze-backend main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/fieldfuze-backend-linux main.go

# Cleanup
clean:
	rm -rf bin/
	go mod tidy

# Docker
docker-build:
	docker build -t fieldfuze-backend .

docker-run:
	docker run -p 8081:8081 fieldfuze-backend

# Database
db-local:
	docker run -p 8000:8000 amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb

# All quality checks
check: fmt lint test

# Setup development environment
setup:
	go mod download
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
```

## Testing

### Test Structure

```
tests/
├── unit/                       # Unit tests
│   ├── controllers/
│   ├── services/
│   ├── repository/
│   └── utils/
├── integration/                # Integration tests
│   ├── api/
│   └── database/
└── e2e/                       # End-to-end tests
    └── scenarios/
```

### Unit Testing

#### Controller Tests
```go
// controller/user_controller_test.go
package controller

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestUserController_Register(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    // Mock dependencies
    mockRepo := &MockUserRepository{}
    mockLogger := &MockLogger{}
    mockJWT := &MockJWTManager{}

    controller := NewUserController(context.Background(), mockRepo, mockLogger, mockJWT)

    // Test data
    user := models.User{
        Email:     "test@example.com",
        Username:  "testuser",
        Password:  "password123",
        FirstName: "Test",
        LastName:  "User",
    }

    jsonData, _ := json.Marshal(user)
    c.Request = httptest.NewRequest("POST", "/user/register", bytes.NewBuffer(jsonData))
    c.Request.Header.Set("Content-Type", "application/json")

    // Mock expectations
    mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(&user, nil)

    // Execute
    controller.Register(c)

    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    mockRepo.AssertExpectations(t)
}
```

#### Service Tests
```go
// services/user_service_test.go
package services

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestUserService_CreateUser(t *testing.T) {
    // Setup
    mockRepo := &MockUserRepository{}
    service := NewUserService(context.Background(), mockRepo)

    // Test data
    user := &models.User{
        Email:     "test@example.com",
        Username:  "testuser",
        FirstName: "Test",
        LastName:  "User",
    }

    // Mock expectations
    mockRepo.On("CreateUser", mock.Anything, user).Return(user, nil)

    // Execute
    result, err := service.CreateUser(user)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, user.Email, result.Email)
    mockRepo.AssertExpectations(t)
}
```

#### Repository Tests
```go
// repository/user_repo_test.go
package repository

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestUserRepository_CreateUser(t *testing.T) {
    // Setup test database
    db := setupTestDB()
    defer teardownTestDB(db)

    repo := NewUserRepository(db, config, logger)

    // Test data
    user := &models.User{
        Email:     "test@example.com",
        Username:  "testuser",
        FirstName: "Test",
        LastName:  "User",
    }

    // Execute
    result, err := repo.CreateUser(context.Background(), user)

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, result.ID)
    assert.Equal(t, user.Email, result.Email)
}
```

### Integration Testing

#### API Integration Tests
```go
// tests/integration/api_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type APITestSuite struct {
    suite.Suite
    server *httptest.Server
    client *http.Client
}

func (suite *APITestSuite) SetupSuite() {
    // Setup test server
    suite.server = setupTestServer()
    suite.client = &http.Client{}
}

func (suite *APITestSuite) TearDownSuite() {
    suite.server.Close()
}

func (suite *APITestSuite) TestUserRegistrationFlow() {
    // Register user
    registerData := map[string]string{
        "email":      "test@example.com",
        "username":   "testuser",
        "password":   "password123",
        "first_name": "Test",
        "last_name":  "User",
    }

    jsonData, _ := json.Marshal(registerData)
    resp, err := suite.client.Post(
        suite.server.URL+"/api/v1/auth/user/register",
        "application/json",
        bytes.NewBuffer(jsonData),
    )

    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

    // Login user
    loginData := map[string]string{
        "email":    "test@example.com",
        "password": "password123",
    }

    jsonData, _ = json.Marshal(loginData)
    resp, err = suite.client.Post(
        suite.server.URL+"/api/v1/auth/user/login",
        "application/json",
        bytes.NewBuffer(jsonData),
    )

    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func TestAPITestSuite(t *testing.T) {
    suite.Run(t, new(APITestSuite))
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific test
go test -v ./controller -run TestUserController_Register

# Run tests in a specific package
go test ./services/

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

## Code Quality

### Linting Configuration

#### .golangci.yml
```yaml
linters-settings:
  govet:
    check-shadowing: true
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
  goconst:
    min-len: 2
    min-occurrences: 2

linters:
  enable:
    - bodyclose
    - deadcode
    - depguard
    - dogsled
    - dupl
    - errcheck
    - exportloopref
    - exhaustive
    - funlen
    - gochecknoinits
    - goconst
    - gocritic
    - gocyclo
    - gofmt
    - goimports
    - gomnd
    - goprintffuncname
    - gosec
    - gosimple
    - govet
    - ineffassign
    - lll
    - misspell
    - nakedret
    - noctx
    - nolintlint
    - staticcheck
    - structcheck
    - stylecheck
    - typecheck
    - unconvert
    - unparam
    - unused
    - varcheck
    - whitespace

run:
  timeout: 5m
  tests: false

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
        - funlen
```

### Code Formatting

```bash
# Format all Go files
go fmt ./...

# Fix imports
goimports -w .

# Run linter
golangci-lint run

# Fix auto-fixable issues
golangci-lint run --fix
```

### Pre-commit Hooks

#### .pre-commit-config.yaml
```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-merge-conflict
      - id: check-yaml

  - repo: local
    hooks:
      - id: go-fmt
        name: go-fmt
        entry: gofmt
        language: system
        args: [-l, -s, -w]
        files: \.go$

      - id: go-imports
        name: go-imports
        entry: goimports
        language: system
        args: [-w]
        files: \.go$

      - id: golangci-lint
        name: golangci-lint
        entry: golangci-lint
        language: system
        args: [run]
        files: \.go$
```

## Debugging

### Debug Configuration

#### VS Code launch.json
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Server",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/main.go",
            "env": {
                "APP_ENV": "development",
                "LOG_LEVEL": "debug"
            },
            "args": []
        },
        {
            "name": "Debug Tests",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}",
            "args": ["-test.v"]
        }
    ]
}
```

### Logging for Debugging

```go
// Add debug logging in your code
func (s *UserService) CreateUser(user *models.User) (*models.User, error) {
    s.logger.Debugf("Creating user with email: %s", user.Email)
    
    result, err := s.repo.CreateUser(context.Background(), user)
    if err != nil {
        s.logger.Errorf("Failed to create user: %v", err)
        return nil, err
    }
    
    s.logger.Debugf("User created successfully with ID: %s", result.ID)
    return result, nil
}
```

### Performance Profiling

```go
// Add profiling to main.go for development
import _ "net/http/pprof"

func main() {
    if os.Getenv("APP_ENV") == "development" {
        go func() {
            log.Println(http.ListenAndServe("localhost:6060", nil))
        }()
    }
    
    // ... rest of main function
}
```

Access profiling data:
```bash
# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## Contributing

### Pull Request Process

1. **Fork & Clone**
   ```bash
   git clone https://github.com/your-username/fieldfuze-backend.git
   cd fieldfuze-backend
   ```

2. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make Changes**
   - Follow coding standards
   - Add tests for new functionality
   - Update documentation

4. **Quality Checks**
   ```bash
   make check  # runs fmt, lint, test
   ```

5. **Commit Changes**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

6. **Push & Create PR**
   ```bash
   git push origin feature/your-feature-name
   # Create pull request on GitHub
   ```

### Code Review Guidelines

- **Code Quality**: Follow Go best practices
- **Testing**: Include unit and integration tests
- **Documentation**: Update relevant documentation
- **Breaking Changes**: Clearly mark and explain
- **Performance**: Consider performance implications

### Release Process

1. **Version Bump**
   ```bash
   # Update version in config and documentation
   git tag v1.0.1
   git push origin v1.0.1
   ```

2. **Changelog**
   - Update CHANGELOG.md
   - Document breaking changes
   - List new features and fixes

3. **Release Notes**
   - Create GitHub release
   - Include migration notes if needed

## Troubleshooting

### Common Issues

#### DynamoDB Connection Issues
```bash
# Check if local DynamoDB is running
curl http://localhost:8000/

# Check AWS credentials
aws sts get-caller-identity

# Test table creation
aws dynamodb list-tables --endpoint-url http://localhost:8000
```

#### JWT Token Issues
```bash
# Decode JWT token for debugging
echo "YOUR_JWT_TOKEN" | base64 -d

# Check token expiration
# Use online JWT decoder or write a small Go program
```

#### Build Issues
```bash
# Clean and rebuild
go clean -cache
go mod tidy
go mod download
go build
```

#### Test Failures
```bash
# Run tests with verbose output
go test -v ./...

# Run specific failing test
go test -v ./controller -run TestUserController_Register

# Check test coverage
go test -cover ./...
```

### Debug Commands

```bash
# Check Go version and environment
go version
go env

# Verify module dependencies
go mod verify
go mod graph

# Check for unused dependencies
go mod tidy

# Analyze binary size
go build -ldflags="-s -w" -o fieldfuze-backend main.go
ls -lh fieldfuze-backend
```

### Getting Help

- **Documentation**: Check the `docs/` directory
- **Issues**: Create GitHub issue with reproduction steps
- **Discussions**: Use GitHub discussions for questions
- **Code Review**: Request review from maintainers

---

**Related Documentation**: [Architecture](ARCHITECTURE.md) | [API Reference](API.md) | [Testing](TESTING.md) | [Deployment](DEPLOYMENT.md)