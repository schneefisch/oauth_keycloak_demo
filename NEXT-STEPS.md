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
| Keycloak 26.5.2 Upgrade | ✅ Complete | High |
| Event CRUD Operations | ❌ Missing | High |
| Role-based Access Control (RBAC) | ❌ Missing | High |
| Multi-Organization Support | ❌ Missing | Medium |
| Token Refresh | ❌ Missing | Medium |
| Frontend Tests | ❌ Missing | Low |
| API Documentation | ❌ Missing | Low |

---

## Phase 0: Upgrade Keycloak to 26.5.2

Keycloak 26.5.2 introduces the native **Organizations** feature, which provides built-in multi-tenancy support. This replaces the need for custom group-based organization management.

### 0.1 Update Docker Compose

**File:** `docker-compose.yml`

Change Keycloak image version:
```yaml
keycloak:
  image: quay.io/keycloak/keycloak:26.5.2
  # ... rest of config
  environment:
    # ... existing env vars
    - KC_FEATURES=scripts,organization  # Enable organization feature
  entrypoint: /opt/keycloak/bin/kc.sh start-dev --import-realm --features=scripts,organization
```

### 0.2 Enable Organizations Feature

The Organizations feature must be enabled via:
- Feature flag: `--features=organization`
- Or environment variable: `KC_FEATURES=organization`

### 0.3 Verify Upgrade

After upgrading:
1. Run `mage build && mage start`
2. Access Keycloak admin console at http://localhost:8081
3. Verify "Organizations" appears in the left sidebar under the realm
4. Confirm existing realm import still works

### 0.4 Key Organization Feature Capabilities

| Capability | Description |
|------------|-------------|
| Organization Management | Create/manage organizations in Keycloak Admin UI |
| Member Management | Add/remove users from organizations |
| Organization Roles | Define roles within organization context |
| Token Claims | Organization membership included in tokens |
| Identity Brokering | Organization-specific identity providers |
| Domain Verification | Link email domains to organizations |

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

## Phase 3: Multi-Organization Support (using Keycloak Organizations)

The README describes organizations (soccer associations) where users can only see their organization's events. With Keycloak 26.5.2, we use the native **Organizations** feature instead of groups.

### 3.1 Keycloak: Create Organizations

In Keycloak Admin Console (http://localhost:8081):

1. Navigate to **Organizations** in the left sidebar
2. Click **Create organization**
3. Create organizations for each soccer association:
   - Name: "FC Example Soccer Club"
   - Alias: "fc-example" (used in API/tokens)
   - Description: Optional
4. Repeat for each association

### 3.2 Keycloak: Add Members to Organizations

For each organization:
1. Go to **Organizations** → Select organization → **Members**
2. Click **Add member**
3. Search and add users
4. Assign organization-specific roles if needed

### 3.3 Keycloak: Configure Organization Token Claims

Ensure organization membership is included in tokens:

**File:** `data/import/events-realm.json`

Add a protocol mapper to include organization in tokens:
```json
{
  "name": "organization",
  "protocol": "openid-connect",
  "protocolMapper": "oidc-organization-membership-mapper",
  "consentRequired": false,
  "config": {
    "id.token.claim": "true",
    "access.token.claim": "true",
    "claim.name": "organization",
    "userinfo.token.claim": "true"
  }
}
```

The token will include an `organization` claim with the user's organization membership:
```json
{
  "organization": {
    "fc-example": {
      "name": "FC Example Soccer Club",
      "roles": ["member"]
    }
  }
}
```

### 3.4 Database: Add Organization Reference to Events

**File:** `data/db/02-add-organization-to-events.sql` (new file)

```sql
-- Add organization_id to events (references Keycloak organization alias)
ALTER TABLE events.events
ADD COLUMN organization_id VARCHAR(255);

-- Create index for organization queries
CREATE INDEX idx_events_organization ON events.events(organization_id);

-- Update existing events with a default organization (optional)
-- UPDATE events.events SET organization_id = 'fc-example' WHERE organization_id IS NULL;
```

Note: We store the Keycloak organization alias (e.g., "fc-example") rather than maintaining a separate organizations table. Keycloak is the source of truth for organization data.

### 3.5 Backend: Extract Organization from Token

**File:** `backend/internal/handlers/auth.go`

Update token validation to extract organization claims:
```go
type TokenInfo struct {
    Active       bool
    Scope        string
    Username     string
    Roles        []string
    Organization map[string]OrganizationMembership  // Add organization extraction
}

type OrganizationMembership struct {
    Name  string   `json:"name"`
    Roles []string `json:"roles"`
}

// In validateToken(), parse the organization claim from introspection response
```

### 3.6 Backend: Organization-Scoped Queries

**File:** `backend/internal/repository/events.go`

Add organization-scoped methods to the interface:
```go
type EventsRepository interface {
    GetEvents(ctx context.Context) (models.Events, error)
    GetEventByID(ctx context.Context, id string) (*models.Event, error)
    // New organization-scoped methods
    GetEventsByOrganization(ctx context.Context, orgID string) (models.Events, error)
    CreateEvent(ctx context.Context, event *models.Event) error  // includes org_id
}
```

**File:** `backend/internal/repository/postgres_events.go`

```go
func (r *PostgresEventsRepository) GetEventsByOrganization(ctx context.Context, orgID string) (models.Events, error) {
    query := `SELECT id, date, title, description, location, organization_id
              FROM events.events
              WHERE organization_id = $1
              ORDER BY date DESC`
    // ... implementation
}
```

### 3.7 Backend: Enforce Organization Isolation

**File:** `backend/internal/handlers/events.go`

Modify handlers to filter by user's organization:
```go
func (h *EventsHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
    // Get organization from request context (set by auth middleware)
    orgID := r.Context().Value("organization").(string)

    // Fetch only events for this organization
    events, err := h.repo.GetEventsByOrganization(r.Context(), orgID)
    // ...
}
```

### 3.8 Frontend: Display Organization Context

**File:** `frontend/html/js/services.js`

Extract and store organization from token:
```javascript
// In AuthService, after token exchange
this.organization = this.parseOrganizationFromToken(tokenResponse.access_token);

parseOrganizationFromToken: function(accessToken) {
    const payload = JSON.parse(atob(accessToken.split('.')[1]));
    return payload.organization || {};
}
```

**File:** `frontend/html/templates/events-list.html`

Display organization name in UI:
```html
<div class="organization-header" ng-if="currentOrganization">
    <h2>{{ currentOrganization.name }} Events</h2>
</div>
```

### 3.9 System Admin: Cross-Organization Access

For system admins who need to see all organizations:
- Check for `system-admin` role in token
- If present, skip organization filtering
- Optionally allow organization switching in UI

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

1. **Phase 0** (Keycloak Upgrade) - Must be done first to enable Organizations feature
2. **Phase 1** (Event CRUD) - Foundation for all other features
3. **Phase 2** (RBAC) - Required before multi-org to control who can do what
4. **Phase 4** (Token Refresh) - Improves UX, can be done in parallel
5. **Phase 3** (Multi-Org) - Most complex, requires CRUD, RBAC, and Keycloak Organizations
6. **Phase 5** (Testing/Docs) - Ongoing throughout development

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
| Keycloak Upgrade | `docker-compose.yml` (image version + KC_FEATURES) |
| CRUD Endpoints | `handlers/events.go`, `handlers/routes.go`, `repository/*.go` |
| RBAC | `handlers/auth.go`, `data/import/events-realm.json` |
| Multi-Org (Keycloak) | `data/import/events-realm.json` (org mapper), Keycloak Admin UI |
| Multi-Org (Backend) | `handlers/auth.go`, `handlers/events.go`, `repository/*.go`, `data/db/02-*.sql` |
| Multi-Org (Frontend) | `frontend/html/js/services.js`, `frontend/html/templates/*.html` |
| Token Refresh | `frontend/html/js/services.js` |
| Frontend UI | `frontend/html/templates/*.html`, `frontend/html/js/controllers.js` |
| Tests | `backend/internal/handlers/*_test.go`, `frontend/test/` |
| API Docs | `docs/api/openapi.yaml` |
