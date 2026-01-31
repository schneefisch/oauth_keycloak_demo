# CLAUDE.md

Sports Community Management App - OAuth 2.0/PKCE with Keycloak, Go backend, AngularJS frontend.

## Commands
```bash
mage            # build + start all
mage build      # docker images
mage start/stop # services
mage test       # go test ./...
mage logs/clean
```

## URLs & Credentials
| Service  | URL                  | Credentials              |
|----------|----------------------|--------------------------|
| Frontend | localhost:80         | -                        |
| Backend  | localhost:8082       | -                        |
| Keycloak | localhost:8081       | admin / bad-password     |
| Postgres | localhost:5432       | admin / Q6uktXCjQ        |

**Test User:** `f.roeser+demo@smight.com` / `Test1234567890!`

## Structure
```
backend/
├── cmd/api/main.go          # entry point
└── internal/
    ├── config/              # koanf config
    ├── handlers/            # HTTP + auth middleware
    ├── models/              # Event struct
    └── repository/          # interface + postgres/mock impl
frontend/
├── html/                    # js/, css/, templates/
├── nginx.conf
data/
├── db/                      # SQL init scripts
└── import/                  # Keycloak realm JSON
```

## Stack
Go 1.24 | AngularJS | PostgreSQL 13 | Keycloak 26.5.2 | Nginx | Mage | koanf v2

## Go Conventions

**Module:** `github.com/schneefisch/oauth_keycloak_demo/backend`

**Imports:** Always absolute from module path
```go
import "github.com/schneefisch/oauth_keycloak_demo/backend/internal/handlers"
```

**Routing:** Go 1.22+ stdlib with path vars: `http.HandleFunc("/events/{id}", handler.GetEventByID)`

**Repository Pattern:** Interface-based data access
```go
type EventsRepository interface {
    GetEvents(ctx context.Context) (models.Events, error)
    GetEventByID(ctx context.Context, id string) (*models.Event, error)
}
```

**Env Vars:** `PORT`, `KEYCLOAK_URL`, `DB_HOST/PORT/USER/PASSWORD/NAME`, `CLIENT_ID/SECRET`, `REQUIRED_SCOPE`, `REALM_NAME`

## Auth Flow
1. Frontend → Keycloak (PKCE) → access token
2. Request with `Authorization: Bearer <token>`
3. Backend → Keycloak introspection → validate + check scope

## API
| Method | Path         | Auth | Description     |
|--------|--------------|------|-----------------|
| GET    | /events      | Yes  | List events     |
| GET    | /events/{id} | Yes  | Get event by ID |
| GET    | /health      | No   | Health check    |

## Database
Schema: `events` | Table: `events.events`
```sql
id VARCHAR(36) PK, date TIMESTAMP, title VARCHAR(255), description TEXT, location VARCHAR(255)
```

## Testing
- Tests: `*_test.go` alongside source
- Mocks: `repository/mock_events.go`, interface-based
- Run: `cd backend && go test ./...`

## CI/CD Verification (Required)

**GitHub Actions runs on every push.** Before marking any code change as complete:

1. **Push changes** to the feature branch
2. **Check CI status** using: `gh run list --branch <branch> --limit 1`
3. **View details if failed**: `gh run view <run-id> --log-failed`
4. **Fix failures** and push again until CI passes
5. **Only then** consider the task complete

**CI checks performed:**
- `go test -race` - Unit tests with race detection
- `go vet` - Static analysis
- `gofmt` - Code formatting
- `staticcheck` - Additional linting
- Docker build - Image builds successfully

**Quick CI commands:**
```bash
gh run list --branch $(git branch --show-current) --limit 3  # Recent runs
gh run watch                                                  # Watch current run
gh run view --log-failed                                      # See failures
```

## Key Files
| Purpose        | Path                                    |
|----------------|-----------------------------------------|
| Entry point    | backend/cmd/api/main.go                 |
| Config         | backend/internal/config/config.go       |
| Auth middleware| backend/internal/handlers/auth.go       |
| Routes         | backend/internal/handlers/routes.go     |
| Handlers       | backend/internal/handlers/events.go     |
| Model          | backend/internal/models/event.go        |
| Repository     | backend/internal/repository/events.go   |
| Docker         | docker-compose.yml                      |
| Build          | Magefile.go                             |
| CI Pipeline    | .github/workflows/ci.yml                |

## Common Tasks
- **New endpoint:** handler in `handlers/` → register in `routes.go` → add tests
- **New model:** struct in `models/` → update repository interface → implement postgres + mock
- **Keycloak:** edit `data/import/events-realm.json` → restart
- **DB migration:** add numbered SQL to `data/db/`

## Troubleshooting
- Token fails: check `KEYCLOAK_URL=http://keycloak:8080` in Docker
- DB connection: use `DB_HOST=postgres` in Docker
- CORS: auth middleware sets `Access-Control-Allow-Origin: *`
