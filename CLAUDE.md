# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go monorepo structured as a **Go workspace** using Go 1.25.0. It contains microservices and shared packages for the SSE-Wiki backend system. The codebase is primarily documented in Chinese.

## Repository Structure

The repository follows a monorepo pattern with Go workspaces (`go.work`):

- **`services/`** - Microservices (e.g., `sse-wiki`, `auth-service`)
- **`packages/`** - Shared packages used across services (e.g., `auth-sdk`, `response`)

### Standard Service Structure

Each service follows this Go project layout:

- **`cmd/`** - Executable entry points (e.g., `cmd/server/main.go`)
- **`config/`** - Configuration management
- **`internal/`** - Internal code not importable by other packages
  - `dto/` - Data Transfer Objects for request/response
  - `handler/` - HTTP request handlers (Gin framework)
  - `middleware/` - Custom middlewares (auth, logging, etc.)
  - `model/` - Database models (GORM)
  - `repository/` - Data access layer (DAO)
  - `route/` - Route definitions
  - `service/` - Business logic layer
- **`pkg/`** - Public utilities exportable to other packages (optional)

**Important**: Services should not export functionality; they import from shared packages.

## Commands

### Root Directory

From the repository root, you can operate on any service or package:

```sh
# Install dependencies for a specific service/package
make install <service-or-package-name>

# Run a specific service/package
make run <service-or-package-name>

# Build a specific service/package
make build <service-or-package-name>

# Clean all build artifacts
make clean
```

Examples:
```sh
make install sse-wiki
make run sse-wiki
make build auth-service
```

### Within a Service/Package Directory

Navigate to a service directory (e.g., `services/sse-wiki/`) and run:

```sh
# Install dependencies
make install
# or manually:
go mod tidy

# Run the service
make run
# or manually:
go run cmd/server/main.go
```

## Architecture

### Technology Stack

- **Framework**: Gin (HTTP web framework)
- **Database**: PostgreSQL with GORM ORM
- **Config**: koanf (supports YAML files + environment variables)
- **CORS**: gin-contrib/cors

### Configuration System

Configuration uses **koanf** which supports:
- YAML configuration files (e.g., `config.yaml`)
- Environment variable overrides with `APP_` prefix
  - Example: `APP_DATABASE_HOST` overrides `database.host` in config

The config system is located in the `config/` package and follows a singleton pattern.

### Service Architecture (Layered Architecture)

The main service entry point is `cmd/server/main.go`:
1. Loads configuration from `config.yaml`
2. Initializes database connection using `packages/database`
3. Auto-migrates database tables
4. Sets up routes via `route.SetupRouter(db)`
5. Starts HTTP server

**Request Flow** (Layered Architecture):
```
HTTP Request
    ↓
Route (路由层) - URL mapping
    ↓
Middleware (中间件层) - Auth, logging, CORS
    ↓
Handler (处理层) - Request validation, call service
    ↓
Service (业务层) - Business logic, transaction control
    ↓
Repository (数据层) - Database operations
    ↓
Model (模型层) - GORM models
    ↓
Database (PostgreSQL/Redis)
```

**Layer Responsibilities**:

- **Route** (`internal/route/`) - Define HTTP endpoints, bind handlers, apply middlewares
- **Middleware** (`internal/middleware/`) - Authentication, authorization, logging, error recovery
- **Handler** (`internal/handler/`) - Parse request, validate parameters, call service, return response
- **Service** (`internal/service/`) - Core business logic, business rule validation, transaction management
- **Repository** (`internal/repository/`) - Data access layer, CRUD operations, database queries
- **Model** (`internal/model/`) - GORM models, database table definitions
- **DTO** (`internal/dto/`) - Request/response data structures

### Shared Packages (packages/)

All shared packages use the import prefix `terminal-terrace/`.

**`database` package**: Unified database connection management
- `InitPostgres(config)` - Initialize PostgreSQL connection with GORM
- `InitRedis(config)` - Initialize Redis connection
- Connection pool configuration
- **Important**: Only provides connections, does NOT contain business SQL
- Business queries should be in each service's `repository` layer

**`response` package**: Standardized API response format
- `SuccessResponse(data)` - Returns success response with code 100
- `ErrorResponse(code, message)` - Returns error response
- `BusinessError` - Business error wrapper
- Response structure: `{ "code": int, "message": string, "data": any }`

**`auth-sdk` package**: Authentication middleware
- JWT authentication middleware (planned)
- Used across all services for unified auth

### Database Connection Pattern

**✅ Recommended Approach** (Using `packages/database`):

```go
// cmd/server/main.go
import "terminal-terrace/database"

func main() {
    config.Load("config.yaml")

    // Initialize database using shared package
    db, err := database.InitPostgres(&database.PostgresConfig{
        Username:        config.Conf.Database.Username,
        Password:        config.Conf.Database.Password,
        Host:            config.Conf.Database.Host,
        Port:            config.Conf.Database.Port,
        Database:        config.Conf.Database.Database,
        SSLMode:         config.Conf.Database.SSLMode,
        LogLevel:        config.Conf.Log.Level,
        MaxIdleConns:    config.Conf.Database.MaxIdleConns,
        MaxOpenConns:    config.Conf.Database.MaxOpenConns,
        ConnMaxLifetime: time.Duration(config.Conf.Database.MaxLifetime) * time.Second,
    })

    // Auto-migrate tables
    model.InitTable(db)

    // Pass db to router
    router := route.SetupRouter(db)
    router.Run(":8080")
}
```

**Repository Layer** (Business SQL):

```go
// internal/repository/user_repository.go
type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByID(id uint) (*model.User, error) {
    var user model.User
    err := r.db.First(&user, id).Error
    return &user, err
}
```

**Key Principles**:
- Database connection initialized ONCE in `main.go` using `packages/database`
- Connection passed down through layers via dependency injection
- Business SQL queries are in service's `repository` layer
- Connection pooling configured: 10 idle, 100 max open connections (configurable)

### CORS Configuration

CORS is configured in the router setup with frontend URL from environment variable `FRONTEND_URL` (defaults to `http://localhost:5173`).

## Module Names and Imports

All Go modules use the import prefix `terminal-terrace/`:

**Services**:
- `terminal-terrace/sse-wiki`
- `terminal-terrace/auth-service`
- `terminal-terrace/template`

**Shared Packages**:
- `terminal-terrace/database`
- `terminal-terrace/response`
- `terminal-terrace/auth-sdk`

**Importing shared packages**:
```go
import (
    "terminal-terrace/database"
    "terminal-terrace/response"
)
```

**go.mod setup** (in each service):
```go
module terminal-terrace/your-service

require (
    terminal-terrace/database v0.0.0
    terminal-terrace/response v0.0.0
)

replace (
    terminal-terrace/database => ../../packages/database
    terminal-terrace/response => ../../packages/response
)
```

## Development Workflow

When creating a new service, follow these steps:

1. **Copy the template**: Use `services/template/` as the starting point
2. **Initialize database**: Use `packages/database` for connection
3. **Define models**: Create GORM models in `internal/model/`
4. **Define DTOs**: Create request/response structures in `internal/dto/`
5. **Implement repository**: Data access layer in `internal/repository/`
6. **Implement service**: Business logic in `internal/service/`
7. **Implement handler**: HTTP handlers in `internal/handler/`
8. **Register routes**: Define routes in `internal/route/`

See `services/template/README.md` for detailed guidance on each layer.
