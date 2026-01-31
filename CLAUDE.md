# CLAUDE.md - AI Assistant Guide for oauth_keycloak_demo

This document provides guidance for AI assistants working with this codebase.

## Project Overview

This is a **Sports Community Management App** demonstrating OAuth 2.0/PKCE authentication with Keycloak. It manages sports events for soccer associations with role-based access control.

### Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Frontend   │────▶│   Backend   │────▶│  PostgreSQL │
│  (AngularJS)│     │    (Go)     │     │             │
└──────┬──────┘     └──────┬──────┘     └─────────────┘
       │                   │
       │                   │ Token introspection
       │                   ▼
       │            ┌─────────────┐
       └───────────▶│  Keycloak   │
          PKCE flow │   (IdP)     │
                    └─────────────┘
```

## Quick Reference

### Build & Run Commands

```bash
# Default: Build and start all services
mage

# Individual commands
mage build     # Build Docker images (backend + frontend)
mage start     # Start all services with docker-compose
mage stop      # Stop all services
mage test      # Run Go tests (cd backend && go test ./...)
mage logs      # View service logs
mage clean     # Remove containers and images
```

### Service URLs (Local)

| Service  | URL                    | Notes                          |
|----------|------------------------|--------------------------------|
| Frontend | http://localhost:80    | AngularJS app via nginx        |
| Backend  | http://localhost:8082  | Go API (mapped from 8080)      |
| Keycloak | http://localhost:8081  | Admin: admin / bad-password    |
| Postgres | localhost:5432         | Admin: admin / Q6uktXCjQ       |

### Test User

- Username: `f.roeser+demo@smight.com`
- Password: `Test1234567890!`

## Project Structure

```
/
├── backend/                    # Go backend service
│   ├── cmd/api/main.go        # Application entry point
│   └── internal/              # Private application code
│       ├── config/            # Configuration (koanf-based)
│       ├── handlers/          # HTTP handlers + auth middleware
│       ├── models/            # Data models (Event)
│       └── repository/        # Data access layer (interface + implementations)
├── frontend/                   # AngularJS frontend
│   ├── html/                  # Static assets (js/, css/, templates/)
│   ├── nginx.conf             # Nginx configuration
│   └── Dockerfile
├── data/
│   ├── db/                    # Database init scripts
│   └── import/                # Keycloak realm configuration
├── docs/                       # Technical documentation
├── bruno_request_collection/   # Bruno API test collection
├── docker-compose.yml          # Service orchestration
└── Magefile.go                 # Build automation
```

## Technology Stack

| Component | Technology        | Version |
|-----------|-------------------|---------|
| Backend   | Go                | 1.24    |
| Frontend  | AngularJS         | 1.x     |
| Web Server| Nginx             | -       |
| Database  | PostgreSQL        | 13      |
| IdP       | Keycloak          | 26.3    |
| Build     | Mage              | -       |
| Config    | koanf             | v2      |

## Code Conventions

### Go Backend

**Module Path**: `github.com/schneefisch/oauth_keycloak_demo/backend`

**Import Style**: Always use absolute imports from module path:
```go
import (
    "github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
    "github.com/schneefisch/oauth_keycloak_demo/backend/internal/handlers"
    "github.com/schneefisch/oauth_keycloak_demo/backend/internal/repository"
)
```

**HTTP Routing**: Uses Go 1.22+ standard library routing with path variables:
```go
http.HandleFunc("/events/{id}", handler.GetEventByID)
```

**Repository Pattern**: Data access abstracted via interfaces:
```go
type EventsRepository interface {
    GetEvents(ctx context.Context) (models.Events, error)
    GetEventByID(ctx context.Context, id string) (*models.Event, error)
}
```

**Configuration**: Environment variables with prefix mapping (e.g., `DB_HOST` → `db.host`):
- `PORT` - Server port
- `KEYCLOAK_URL` - Keycloak base URL
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - Database config
- `CLIENT_ID`, `CLIENT_SECRET`, `REQUIRED_SCOPE`, `REALM_NAME` - Auth config

**Error Handling**: Standard `http.Error()` with appropriate status codes.

### Testing Conventions

**Test Files**: Located alongside source files with `_test.go` suffix.

**Mock Pattern**: Use interface-based mocks:
- `repository/mock_events.go` - Mock repository for testing
- `handlers/auth_test.go` - Mock HTTP client for auth tests

**Test Helpers**: Create helper functions for common setup:
```go
func createMockAuthConfig() config.AuthConfig { ... }
func createMockHTTPClient() *MockHTTPClient { ... }
```

**Running Tests**:
```bash
cd backend && go test ./...
```

### Authentication Flow

1. Frontend uses PKCE flow to obtain access token from Keycloak
2. Access token sent as `Authorization: Bearer <token>` header
3. Backend validates token via Keycloak introspection endpoint
4. Backend checks for required scope in token claims

**Auth Middleware Pattern**:
```go
authMiddleware := NewAuthMiddleware(authConfig)
http.HandleFunc("/events", authMiddleware(handler.GetEvents))
```

## API Endpoints

| Method | Path           | Auth Required | Description           |
|--------|----------------|---------------|-----------------------|
| GET    | /events        | Yes           | List all events       |
| GET    | /events/{id}   | Yes           | Get event by ID       |
| GET    | /health        | No            | Health check          |

## Database Schema

**Schema**: `events` (separate from `keycloak` schema)

**Table**: `events.events`
```sql
CREATE TABLE events.events (
    id VARCHAR(36) PRIMARY KEY,
    date TIMESTAMP NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(255)
);
```

## Docker Development

**Build Images**:
```bash
cd backend && docker build -t backend-service:latest .
cd frontend && docker build -t frontend:latest .
```

**Network**: All services communicate on `app-network` bridge network.

**Volumes**:
- `postgres_data` - Persistent database storage
- `./data/import` → Keycloak realm import
- `./data/db/01-create-events-schema.sql` → DB initialization

## Common Tasks

### Adding a New API Endpoint

1. Define handler method in `backend/internal/handlers/`
2. Register route in `routes.go` with auth middleware if needed
3. Add tests in corresponding `_test.go` file

### Adding a New Model

1. Create model struct in `backend/internal/models/`
2. Update repository interface in `backend/internal/repository/events.go`
3. Implement in `postgres_events.go` and `mock_events.go`

### Modifying Keycloak Configuration

1. Edit `data/import/events-realm.json`
2. Restart Keycloak or use admin console at http://localhost:8081

### Database Migrations

Add SQL scripts to `data/db/` with numbered prefix (e.g., `02-add-new-table.sql`). Scripts run alphabetically on first container start.

## Troubleshooting

**Token validation fails**: Check `KEYCLOAK_URL` env var matches container network (use `http://keycloak:8080` in Docker).

**Database connection issues**: Ensure `DB_HOST=postgres` when running in Docker.

**CORS errors**: Auth middleware sets `Access-Control-Allow-Origin: *` for development.

## Key Files Reference

| Purpose               | File                                      |
|-----------------------|-------------------------------------------|
| App entry point       | `backend/cmd/api/main.go`                 |
| Configuration         | `backend/internal/config/config.go`       |
| Auth middleware       | `backend/internal/handlers/auth.go`       |
| Route setup           | `backend/internal/handlers/routes.go`     |
| Event handlers        | `backend/internal/handlers/events.go`     |
| Event model           | `backend/internal/models/event.go`        |
| Repository interface  | `backend/internal/repository/events.go`   |
| Docker orchestration  | `docker-compose.yml`                      |
| Build automation      | `Magefile.go`                             |
