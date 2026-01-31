# Next Steps - OAuth Keycloak Demo

This document outlines the remaining work to complete the OAuth/Keycloak demo application based on the goals described in the README.

## Current Implementation Status

### Working Features

| Feature | Status | Notes |
|---------|--------|-------|
| OAuth 2.0 Authorization Code Flow | ✅ Complete | Full PKCE S256 implementation |
| Keycloak Integration | ✅ Complete | Realm, clients, and users configured |
| Token Introspection | ✅ Complete | Backend validates tokens via Keycloak |
| Scope-based Access Control | ✅ Complete | `events-api-access` scope required |
| GET /events | ✅ Complete | List all events with auth |
| GET /events/{id} | ✅ Complete | Get single event with auth |
| GET /health | ✅ Complete | Health check endpoint |
| Frontend Login/Logout | ✅ Complete | PKCE flow with Keycloak |
| Event List View | ✅ Complete | Displays events after authentication |
| Event Detail View | ✅ Complete | Shows individual event details |
| Backend Tests | ✅ Complete | ~1000 lines of test coverage |
| Docker Compose Setup | ✅ Complete | All 4 services orchestrated |

### Missing Features (from README goals)

| Feature | Status | Priority |
|---------|--------|----------|
| Event CRUD Operations | ❌ Missing | High |
| Role-based Access Control (RBAC) | ❌ Missing | High |
| Multi-Organization Support | ❌ Missing | Medium |
| Token Refresh | ❌ Missing | Medium |
| Frontend Tests | ❌ Missing | Low |
| API Documentation | ❌ Missing | Low |

---

## Phase 1: Event CRUD Operations

The backend currently only supports read operations. The frontend has stub methods for create/update/delete but the backend doesn't implement them.

### 1.1 Backend: Add Create Event Endpoint

**File:** `backend/internal/handlers/events.go`

```go
// POST /events
func (h *EventsHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
    // Parse request body into models.Event
    // Validate required fields (title, date, location)
    // Generate UUID for new event
    // Call repository.CreateEvent()
    // Return 201 Created with event data
}
```

**Repository changes needed:**
- Add `CreateEvent(ctx context.Context, event *models.Event) error` to interface
- Implement in `postgres_events.go`
- Implement in `mock_events.go`

### 1.2 Backend: Add Update Event Endpoint

**File:** `backend/internal/handlers/events.go`

```go
// PUT /events/{id}
func (h *EventsHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
    // Extract ID from path
    // Parse request body
    // Validate event exists
    // Call repository.UpdateEvent()
    // Return 200 OK with updated event
}
```

**Repository changes needed:**
- Add `UpdateEvent(ctx context.Context, id string, event *models.Event) error` to interface

### 1.3 Backend: Add Delete Event Endpoint

**File:** `backend/internal/handlers/events.go`

```go
// DELETE /events/{id}
func (h *EventsHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
    // Extract ID from path
    // Validate event exists
    // Call repository.DeleteEvent()
    // Return 204 No Content
}
```

**Repository changes needed:**
- Add `DeleteEvent(ctx context.Context, id string) error` to interface

### 1.4 Register New Routes

**File:** `backend/internal/handlers/routes.go`

```go
mux.HandleFunc("POST /events", handler.authMiddleware(handler.CreateEvent))
mux.HandleFunc("PUT /events/{id}", handler.authMiddleware(handler.UpdateEvent))
mux.HandleFunc("DELETE /events/{id}", handler.authMiddleware(handler.DeleteEvent))
```

### 1.5 Frontend: Event Management UI

**Files to create/modify:**
- `frontend/html/templates/event-form.html` - Create/edit form template
- `frontend/html/js/controllers.js` - Add form handling logic
- `frontend/html/templates/events-list.html` - Add create/edit/delete buttons

---

## Phase 2: Role-Based Access Control (RBAC)

The README describes three roles that are not currently enforced:
- **System Admin** - Overall system administrator
- **Organization Maintainer** - Manages association settings and users
- **Organization User** - Regular users (parents) within an association

### 2.1 Keycloak: Configure Roles

**File:** `data/import/events-realm.json`

Add realm roles:
```json
{
  "roles": {
    "realm": [
      { "name": "system-admin", "description": "System administrator" },
      { "name": "org-maintainer", "description": "Organization maintainer" },
      { "name": "org-user", "description": "Organization user" }
    ]
  }
}
```

### 2.2 Backend: Extract Roles from Token

**File:** `backend/internal/handlers/auth.go`

Modify `validateToken()` to extract and return roles:
```go
type TokenInfo struct {
    Active   bool
    Scope    string
    Username string
    Roles    []string  // Add roles extraction
}
```

### 2.3 Backend: Role-Based Middleware

**File:** `backend/internal/handlers/auth.go`

Create role-checking middleware:
```go
func (h *EventsHandler) requireRole(role string, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get token info from context
        // Check if user has required role
        // Return 403 Forbidden if not
        next(w, r)
    }
}
```

### 2.4 Apply Role Requirements

| Endpoint | Required Role |
|----------|---------------|
| GET /events | org-user, org-maintainer, system-admin |
| GET /events/{id} | org-user, org-maintainer, system-admin |
| POST /events | org-maintainer, system-admin |
| PUT /events/{id} | org-maintainer, system-admin |
| DELETE /events/{id} | org-maintainer, system-admin |

---

## Phase 3: Multi-Organization Support

The README describes organizations (soccer associations) where users can only see their organization's events.

### 3.1 Database: Add Organization Tables

**File:** `data/db/02-create-organizations.sql` (new file)

```sql
CREATE TABLE events.organizations (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add organization_id to events
ALTER TABLE events.events
ADD COLUMN organization_id VARCHAR(36) REFERENCES events.organizations(id);

-- Create user-organization mapping (or use Keycloak groups)
CREATE TABLE events.user_organizations (
    user_id VARCHAR(36) NOT NULL,
    organization_id VARCHAR(36) NOT NULL REFERENCES events.organizations(id),
    PRIMARY KEY (user_id, organization_id)
);
```

### 3.2 Keycloak: Configure Groups

Use Keycloak groups to represent organizations:
- Create group per organization
- Assign users to groups
- Include group membership in tokens

### 3.3 Backend: Organization-Scoped Queries

Modify repository to filter events by organization:
```go
GetEventsByOrganization(ctx context.Context, orgID string) (models.Events, error)
```

### 3.4 Backend: Extract Organization from Token

Parse group/organization claims from introspection response and filter data accordingly.

---

## Phase 4: Token Refresh

Currently, tokens expire after 5 minutes and users must re-login.

### 4.1 Frontend: Implement Token Refresh

**File:** `frontend/html/js/services.js`

```javascript
// Add to AuthService
refreshAccessToken: function() {
    // Check if token is about to expire (e.g., < 30 seconds)
    // Call Keycloak token endpoint with refresh_token
    // Update stored tokens
    // Return new access token
}

// Add automatic refresh before API calls
getValidAccessToken: function() {
    if (this.isTokenExpiring()) {
        return this.refreshAccessToken();
    }
    return $q.resolve(this.accessToken);
}
```

### 4.2 Frontend: Add Token Expiry Check

```javascript
isTokenExpiring: function() {
    if (!this.expiresAt) return true;
    // Check if token expires in less than 30 seconds
    return (this.expiresAt - Date.now()) < 30000;
}
```

### 4.3 Update EventsService

Modify API calls to use `getValidAccessToken()` instead of `getAccessToken()`.

---

## Phase 5: Testing & Documentation

### 5.1 Frontend Tests

Set up testing framework:
- Install Jasmine and Karma
- Create test files for services and controllers
- Add `mage test:frontend` command

**Test cases needed:**
- AuthService PKCE flow
- EventsService API calls
- Controller state management
- Error handling

### 5.2 API Documentation

Create OpenAPI/Swagger specification:

**File:** `docs/api/openapi.yaml`

Document all endpoints with:
- Request/response schemas
- Authentication requirements
- Error responses
- Example payloads

### 5.3 Integration Tests

Add end-to-end tests that:
- Start all services
- Perform OAuth flow
- Execute API calls
- Verify responses

---

## Implementation Order Recommendation

1. **Phase 1** (Event CRUD) - Foundation for all other features
2. **Phase 2** (RBAC) - Required before multi-org to control who can do what
3. **Phase 4** (Token Refresh) - Improves UX, can be done in parallel
4. **Phase 3** (Multi-Org) - Most complex, requires CRUD and RBAC first
5. **Phase 5** (Testing/Docs) - Ongoing throughout development

---

## Quick Wins (Can Implement Immediately)

1. **Input Validation** - Add request body validation for future CRUD endpoints
2. **Error Response Standardization** - Create consistent error response format
3. **Logging** - Add structured logging with request IDs
4. **CORS Hardening** - Replace `*` with specific allowed origins
5. **API Versioning** - Add `/api/v1/` prefix to endpoints

---

## File Reference

| Task | Files to Modify |
|------|-----------------|
| CRUD Endpoints | `handlers/events.go`, `handlers/routes.go`, `repository/*.go` |
| RBAC | `handlers/auth.go`, `data/import/events-realm.json` |
| Multi-Org | `data/db/*.sql`, `models/`, `repository/`, `handlers/` |
| Token Refresh | `frontend/html/js/services.js` |
| Frontend UI | `frontend/html/templates/*.html`, `frontend/html/js/controllers.js` |
| Tests | `backend/internal/handlers/*_test.go`, `frontend/test/` |
| API Docs | `docs/api/openapi.yaml` |
