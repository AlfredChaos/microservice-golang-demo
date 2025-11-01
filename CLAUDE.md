# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Run
```bash
# Build all services (auto-generates protobuf and swagger)
make build

# Build specific service
make build-api-gateway
make build-user-service
make build-book-service
make build-nice-service

# Run services (requires separate terminals)
make run-gateway    # Runs API Gateway on port 8080
make run-user       # Runs User Service
make run-book       # Runs Book Service
make run-nice       # Runs Nice Service (RabbitMQ consumer)
```

### Code Generation
```bash
# Generate protobuf code from .proto files
make proto

# Generate Swagger documentation from annotations
make swagger

# Install required development tools
make install-tools
```

### Dependencies
```bash
# Download and tidy Go modules
make deps
```

### Testing
```bash
# Run tests (standard Go testing)
go test ./...

# Run tests for specific module
go test ./internal/api-gateway/...
go test ./internal/user-service/...
```

## Architecture Overview

This is a microservices demo with gRPC inter-service communication and HTTP REST API via an API Gateway.

### Core Services
- **API Gateway** (`cmd/api-gateway/`): HTTP BFF layer using Gin, provides REST endpoints and Swagger UI
- **User Service** (`cmd/user-service/`): gRPC service for user-related operations
- **Book Service** (`cmd/book-service/`): gRPC service for book-related operations
- **Nice Service** (`cmd/nice-service/`): RabbitMQ consumer service

### Communication Patterns
- **Synchronous**: API Gateway → gRPC services (user-service, book-service)
- **Asynchronous**: API Gateway → RabbitMQ → nice-service

### Technology Stack
- **HTTP Framework**: Gin with Swagger/OpenAPI documentation
- **RPC**: gRPC with Protocol Buffers
- **Databases**: MongoDB (primary), Redis (caching)
- **Message Queue**: RabbitMQ
- **Configuration**: Viper with YAML configs
- **Logging**: Uber Zap
- **Dependency Injection**: Manual wire patterns in `internal/*/wire/`

### Project Structure
```
api/           # Protocol Buffer definitions (.proto files)
cmd/           # Service entry points (main.go)
configs/       # YAML configuration files for each service
internal/      # Private application code per service
pkg/           # Shared libraries across services
scripts/       # Code generation scripts
docs/          # Auto-generated Swagger documentation
```

### Service Architecture (Internal)
Each service follows a layered architecture:
- **controller**: HTTP handlers (API Gateway) or gRPC handlers (services)
- **service**: Business logic layer implementing domain interfaces
- **data**: Repository pattern with MongoDB/Redis implementations
- **domain**: Business entities and interfaces
- **client**: External service client management (API Gateway)
- **wire**: Dependency injection setup

### Configuration Management
- Service configs in `configs/` directory (e.g., `api-gateway.yaml`)
- Uses Viper for configuration loading
- Separate configs for different environments (e.g., `api-gateway.prod.yaml`)

### Code Generation Workflow
1. Add protobuf definitions to `api/` directory
2. Run `make proto` to generate gRPC code
3. Add Swagger annotations to controllers and DTOs
4. Run `make swagger` to generate API documentation
5. Build services with `make build`

### Development Notes
- All services use structured logging with Zap
- gRPC connections managed centrally via `pkg/grpcclient`
- MongoDB connections managed via `pkg/db/mongo.go`
- Redis connections managed via `pkg/cache/redis.go`
- RabbitMQ utilities in `pkg/mq/`
- API Gateway automatically serves Swagger UI at `/swagger/index.html`

### Port Configuration
- API Gateway: 8080 (HTTP + Swagger UI)
- User Service: gRPC port (check configs/user-service.yaml)
- Book Service: gRPC port (check configs/book-service.yaml)
- Nice Service: No HTTP port (RabbitMQ consumer only)