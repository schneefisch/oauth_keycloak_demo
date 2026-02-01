package oauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthClaims_HasScope(t *testing.T) {
	tests := []struct {
		name   string
		scopes []string
		scope  string
		want   bool
	}{
		{
			name:   "has scope",
			scopes: []string{"events:read", "events:write"},
			scope:  "events:read",
			want:   true,
		},
		{
			name:   "does not have scope",
			scopes: []string{"events:read"},
			scope:  "events:write",
			want:   false,
		},
		{
			name:   "empty scopes",
			scopes: []string{},
			scope:  "events:read",
			want:   false,
		},
		{
			name:   "nil scopes",
			scopes: nil,
			scope:  "events:read",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &AuthClaims{Scopes: tt.scopes}
			if got := claims.HasScope(tt.scope); got != tt.want {
				t.Errorf("HasScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthClaims_HasRole(t *testing.T) {
	tests := []struct {
		name  string
		roles []string
		role  string
		want  bool
	}{
		{
			name:  "has role",
			roles: []string{"admin", "user"},
			role:  "admin",
			want:  true,
		},
		{
			name:  "does not have role",
			roles: []string{"user"},
			role:  "admin",
			want:  false,
		},
		{
			name:  "empty roles",
			roles: []string{},
			role:  "admin",
			want:  false,
		},
		{
			name:  "nil roles",
			roles: nil,
			role:  "admin",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &AuthClaims{Roles: tt.roles}
			if got := claims.HasRole(tt.role); got != tt.want {
				t.Errorf("HasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthClaims_HasAnyScope(t *testing.T) {
	tests := []struct {
		name           string
		claimsScopes   []string
		requiredScopes []string
		want           bool
	}{
		{
			name:           "has one of required scopes",
			claimsScopes:   []string{"events:read"},
			requiredScopes: []string{"events:read", "events:write"},
			want:           true,
		},
		{
			name:           "has all required scopes",
			claimsScopes:   []string{"events:read", "events:write"},
			requiredScopes: []string{"events:read", "events:write"},
			want:           true,
		},
		{
			name:           "has none of required scopes",
			claimsScopes:   []string{"other:read"},
			requiredScopes: []string{"events:read", "events:write"},
			want:           false,
		},
		{
			name:           "empty required scopes returns true",
			claimsScopes:   []string{"events:read"},
			requiredScopes: []string{},
			want:           true,
		},
		{
			name:           "nil required scopes returns true",
			claimsScopes:   []string{"events:read"},
			requiredScopes: nil,
			want:           true,
		},
		{
			name:           "empty claims scopes",
			claimsScopes:   []string{},
			requiredScopes: []string{"events:read"},
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &AuthClaims{Scopes: tt.claimsScopes}
			if got := claims.HasAnyScope(tt.requiredScopes...); got != tt.want {
				t.Errorf("HasAnyScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthClaims_HasAnyRole(t *testing.T) {
	tests := []struct {
		name          string
		claimsRoles   []string
		requiredRoles []string
		want          bool
	}{
		{
			name:          "has one of required roles",
			claimsRoles:   []string{"user"},
			requiredRoles: []string{"admin", "user"},
			want:          true,
		},
		{
			name:          "has all required roles",
			claimsRoles:   []string{"admin", "user"},
			requiredRoles: []string{"admin", "user"},
			want:          true,
		},
		{
			name:          "has none of required roles",
			claimsRoles:   []string{"guest"},
			requiredRoles: []string{"admin", "user"},
			want:          false,
		},
		{
			name:          "empty required roles returns true",
			claimsRoles:   []string{"user"},
			requiredRoles: []string{},
			want:          true,
		},
		{
			name:          "nil required roles returns true",
			claimsRoles:   []string{"user"},
			requiredRoles: nil,
			want:          true,
		},
		{
			name:          "empty claims roles",
			claimsRoles:   []string{},
			requiredRoles: []string{"admin"},
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &AuthClaims{Roles: tt.claimsRoles}
			if got := claims.HasAnyRole(tt.requiredRoles...); got != tt.want {
				t.Errorf("HasAnyRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthClaims_HasAllScopes(t *testing.T) {
	tests := []struct {
		name           string
		claimsScopes   []string
		requiredScopes []string
		want           bool
	}{
		{
			name:           "has all required scopes",
			claimsScopes:   []string{"events:read", "events:write"},
			requiredScopes: []string{"events:read", "events:write"},
			want:           true,
		},
		{
			name:           "has more than required scopes",
			claimsScopes:   []string{"events:read", "events:write", "events:delete"},
			requiredScopes: []string{"events:read", "events:write"},
			want:           true,
		},
		{
			name:           "missing one required scope",
			claimsScopes:   []string{"events:read"},
			requiredScopes: []string{"events:read", "events:write"},
			want:           false,
		},
		{
			name:           "empty required scopes returns true",
			claimsScopes:   []string{"events:read"},
			requiredScopes: []string{},
			want:           true,
		},
		{
			name:           "nil required scopes returns true",
			claimsScopes:   []string{"events:read"},
			requiredScopes: nil,
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &AuthClaims{Scopes: tt.claimsScopes}
			if got := claims.HasAllScopes(tt.requiredScopes...); got != tt.want {
				t.Errorf("HasAllScopes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthClaims_HasAllRoles(t *testing.T) {
	tests := []struct {
		name          string
		claimsRoles   []string
		requiredRoles []string
		want          bool
	}{
		{
			name:          "has all required roles",
			claimsRoles:   []string{"admin", "user"},
			requiredRoles: []string{"admin", "user"},
			want:          true,
		},
		{
			name:          "has more than required roles",
			claimsRoles:   []string{"admin", "user", "moderator"},
			requiredRoles: []string{"admin", "user"},
			want:          true,
		},
		{
			name:          "missing one required role",
			claimsRoles:   []string{"user"},
			requiredRoles: []string{"admin", "user"},
			want:          false,
		},
		{
			name:          "empty required roles returns true",
			claimsRoles:   []string{"user"},
			requiredRoles: []string{},
			want:          true,
		},
		{
			name:          "nil required roles returns true",
			claimsRoles:   []string{"user"},
			requiredRoles: nil,
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &AuthClaims{Roles: tt.claimsRoles}
			if got := claims.HasAllRoles(tt.requiredRoles...); got != tt.want {
				t.Errorf("HasAllRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSetAuthClaims(t *testing.T) {
	t.Run("set and get claims from context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		claims := &AuthClaims{
			Subject:  "user-123",
			Username: "testuser",
			Email:    "test@example.com",
			Scopes:   []string{"events:read"},
			Roles:    []string{"user"},
		}

		// Set claims in request context
		reqWithClaims := SetAuthClaims(req, claims)

		// Get claims from request context
		gotClaims := GetAuthClaims(reqWithClaims)

		if gotClaims == nil {
			t.Fatal("GetAuthClaims() returned nil")
		}
		if gotClaims.Subject != claims.Subject {
			t.Errorf("Subject = %v, want %v", gotClaims.Subject, claims.Subject)
		}
		if gotClaims.Username != claims.Username {
			t.Errorf("Username = %v, want %v", gotClaims.Username, claims.Username)
		}
		if gotClaims.Email != claims.Email {
			t.Errorf("Email = %v, want %v", gotClaims.Email, claims.Email)
		}
	})

	t.Run("get claims from empty context returns nil", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		gotClaims := GetAuthClaims(req)

		if gotClaims != nil {
			t.Errorf("GetAuthClaims() = %v, want nil", gotClaims)
		}
	})

	t.Run("get claims with wrong type in context returns nil", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		// Put a string instead of AuthClaims in the context
		ctx := context.WithValue(req.Context(), authClaimsKey, "wrong type")
		req = req.WithContext(ctx)

		gotClaims := GetAuthClaims(req)

		if gotClaims != nil {
			t.Errorf("GetAuthClaims() = %v, want nil", gotClaims)
		}
	})
}
