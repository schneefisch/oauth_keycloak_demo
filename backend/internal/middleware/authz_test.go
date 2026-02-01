package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/oauth"
)

func TestAuthzMiddleware_NoClaimsInContext(t *testing.T) {
	config := AuthzConfig{
		RequiredScopes: []string{"events:read"},
	}

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	authzMiddleware := NewAuthzMiddleware(config)
	handler := authzMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
	if handlerCalled {
		t.Error("Handler should not have been called")
	}
}

func TestAuthzMiddleware_MissingRequiredScope(t *testing.T) {
	config := AuthzConfig{
		RequiredScopes: []string{"events:write"},
		RequireAll:     true,
	}

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	authzMiddleware := NewAuthzMiddleware(config)
	handler := authzMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	claims := &oauth.AuthClaims{
		Subject: "user-123",
		Scopes:  []string{"events:read"},
	}
	req = oauth.SetAuthClaims(req, claims)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
	if handlerCalled {
		t.Error("Handler should not have been called")
	}
}

func TestAuthzMiddleware_MissingRequiredRole(t *testing.T) {
	config := AuthzConfig{
		RequiredRoles: []string{"admin"},
		RequireAll:    true,
	}

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	authzMiddleware := NewAuthzMiddleware(config)
	handler := authzMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	claims := &oauth.AuthClaims{
		Subject: "user-123",
		Roles:   []string{"user"},
	}
	req = oauth.SetAuthClaims(req, claims)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
	if handlerCalled {
		t.Error("Handler should not have been called")
	}
}

func TestAuthzMiddleware_HasAllRequiredScopes(t *testing.T) {
	config := AuthzConfig{
		RequiredScopes: []string{"events:read", "events:write"},
		RequireAll:     true,
	}

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	authzMiddleware := NewAuthzMiddleware(config)
	handler := authzMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	claims := &oauth.AuthClaims{
		Subject: "user-123",
		Scopes:  []string{"events:read", "events:write", "events:delete"},
	}
	req = oauth.SetAuthClaims(req, claims)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handlerCalled {
		t.Error("Handler should have been called")
	}
}

func TestAuthzMiddleware_HasAnyRequiredScope_RequireAllFalse(t *testing.T) {
	config := AuthzConfig{
		RequiredScopes: []string{"events:read", "events:write"},
		RequireAll:     false,
	}

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	authzMiddleware := NewAuthzMiddleware(config)
	handler := authzMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	claims := &oauth.AuthClaims{
		Subject: "user-123",
		Scopes:  []string{"events:read"},
	}
	req = oauth.SetAuthClaims(req, claims)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handlerCalled {
		t.Error("Handler should have been called")
	}
}

func TestAuthzMiddleware_HasAllRequiredRoles(t *testing.T) {
	config := AuthzConfig{
		RequiredRoles: []string{"admin", "user"},
		RequireAll:    true,
	}

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	authzMiddleware := NewAuthzMiddleware(config)
	handler := authzMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	claims := &oauth.AuthClaims{
		Subject: "user-123",
		Roles:   []string{"admin", "user", "moderator"},
	}
	req = oauth.SetAuthClaims(req, claims)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handlerCalled {
		t.Error("Handler should have been called")
	}
}

func TestAuthzMiddleware_HasAnyRequiredRole_RequireAllFalse(t *testing.T) {
	config := AuthzConfig{
		RequiredRoles: []string{"admin", "moderator"},
		RequireAll:    false,
	}

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	authzMiddleware := NewAuthzMiddleware(config)
	handler := authzMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	claims := &oauth.AuthClaims{
		Subject: "user-123",
		Roles:   []string{"admin"},
	}
	req = oauth.SetAuthClaims(req, claims)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handlerCalled {
		t.Error("Handler should have been called")
	}
}

func TestAuthzMiddleware_NoRequirements_PassesThrough(t *testing.T) {
	config := AuthzConfig{
		RequiredScopes: []string{},
		RequiredRoles:  []string{},
	}

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	authzMiddleware := NewAuthzMiddleware(config)
	handler := authzMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	claims := &oauth.AuthClaims{
		Subject: "user-123",
	}
	req = oauth.SetAuthClaims(req, claims)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handlerCalled {
		t.Error("Handler should have been called")
	}
}

func TestAuthzMiddleware_CombinedScopesAndRoles_RequireAll(t *testing.T) {
	config := AuthzConfig{
		RequiredScopes: []string{"events:read"},
		RequiredRoles:  []string{"user"},
		RequireAll:     true,
	}

	tests := []struct {
		name           string
		scopes         []string
		roles          []string
		expectedStatus int
		handlerCalled  bool
	}{
		{
			name:           "has both scope and role",
			scopes:         []string{"events:read"},
			roles:          []string{"user"},
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
		{
			name:           "has scope but not role",
			scopes:         []string{"events:read"},
			roles:          []string{"guest"},
			expectedStatus: http.StatusForbidden,
			handlerCalled:  false,
		},
		{
			name:           "has role but not scope",
			scopes:         []string{"events:write"},
			roles:          []string{"user"},
			expectedStatus: http.StatusForbidden,
			handlerCalled:  false,
		},
		{
			name:           "has neither",
			scopes:         []string{"events:write"},
			roles:          []string{"guest"},
			expectedStatus: http.StatusForbidden,
			handlerCalled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			testHandler := func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			}

			authzMiddleware := NewAuthzMiddleware(config)
			handler := authzMiddleware(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			claims := &oauth.AuthClaims{
				Subject: "user-123",
				Scopes:  tt.scopes,
				Roles:   tt.roles,
			}
			req = oauth.SetAuthClaims(req, claims)
			rr := httptest.NewRecorder()

			handler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			if handlerCalled != tt.handlerCalled {
				t.Errorf("Expected handlerCalled=%v, got %v", tt.handlerCalled, handlerCalled)
			}
		})
	}
}

func TestAuthzMiddleware_CombinedScopesAndRoles_RequireAny(t *testing.T) {
	config := AuthzConfig{
		RequiredScopes: []string{"events:read"},
		RequiredRoles:  []string{"admin"},
		RequireAll:     false,
	}

	tests := []struct {
		name           string
		scopes         []string
		roles          []string
		expectedStatus int
		handlerCalled  bool
	}{
		{
			name:           "has both scope and role",
			scopes:         []string{"events:read"},
			roles:          []string{"admin"},
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
		{
			name:           "has scope but not role",
			scopes:         []string{"events:read"},
			roles:          []string{"user"},
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
		{
			name:           "has role but not scope",
			scopes:         []string{"events:write"},
			roles:          []string{"admin"},
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
		{
			name:           "has neither",
			scopes:         []string{"events:write"},
			roles:          []string{"user"},
			expectedStatus: http.StatusForbidden,
			handlerCalled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			testHandler := func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			}

			authzMiddleware := NewAuthzMiddleware(config)
			handler := authzMiddleware(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			claims := &oauth.AuthClaims{
				Subject: "user-123",
				Scopes:  tt.scopes,
				Roles:   tt.roles,
			}
			req = oauth.SetAuthClaims(req, claims)
			rr := httptest.NewRecorder()

			handler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			if handlerCalled != tt.handlerCalled {
				t.Errorf("Expected handlerCalled=%v, got %v", tt.handlerCalled, handlerCalled)
			}
		})
	}
}
