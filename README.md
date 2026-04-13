# Infrastructure

Shared platform and infrastructure components for Go services. This repository provides common utilities and abstractions for building microservices consistently.

## Features

- **Auth**: Keycloak integration and OIDC support.
- **Database**: GORM abstractions for PostgreSQL and MySQL.
- **Messaging**: Background worker support via Asynq (Redis-backed).
- **Server**: Pre-configured Gin-based HTTP server with graceful shutdown.
- **Session**: Redis-backed session management.
- **Utilities**: Structured logging (Logrus), configuration management, and email sending.

## Project Structure

- `/platform`: Core platform components and shared libraries.
- `/common`: Shared data types and utilities used across services.

## Getting Started

### Prerequisites

- Go 1.24+
- Docker (for Redis, PostgreSQL, Keycloak)

### Usage

Add the platform as a dependency in your Go project:

```bash
go get platform.local/platform
```

Note: Since this is likely a local development environment, you may need a `replace` directive in your `go.mod`:

```go
replace platform.local/platform => ../infrastructure/platform
```

### Development

To run tests or build the components:

```bash
# From the platform directory
cd platform
go build ./...
go test ./...
```
