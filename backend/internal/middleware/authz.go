package middleware

import (
	"net/http"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/oauth"
)

// AuthzConfig holds configuration for authorization middleware
type AuthzConfig struct {
	RequiredScopes []string // Scopes required to access the resource
	RequiredRoles  []string // Roles required to access the resource
	RequireAll     bool     // If true, ALL scopes and roles must be present; if false, ANY scope OR role is sufficient
}

// NewAuthzMiddleware creates a new authorization middleware with the given configuration
func NewAuthzMiddleware(config AuthzConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context (set by AuthN middleware)
			claims := oauth.GetAuthClaims(r)
			if claims == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Check authorization based on RequireAll flag
			if !isAuthorized(claims, config) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// User is authorized, call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// isAuthorized checks if the claims meet the authorization requirements
func isAuthorized(claims *oauth.AuthClaims, config AuthzConfig) bool {
	hasRequiredScopes := len(config.RequiredScopes) == 0
	hasRequiredRoles := len(config.RequiredRoles) == 0

	if config.RequireAll {
		// ALL scopes AND ALL roles must be present
		if len(config.RequiredScopes) > 0 {
			hasRequiredScopes = claims.HasAllScopes(config.RequiredScopes...)
		}
		if len(config.RequiredRoles) > 0 {
			hasRequiredRoles = claims.HasAllRoles(config.RequiredRoles...)
		}
		return hasRequiredScopes && hasRequiredRoles
	}

	// ANY scope OR ANY role is sufficient
	if len(config.RequiredScopes) > 0 {
		hasRequiredScopes = claims.HasAnyScope(config.RequiredScopes...)
	}
	if len(config.RequiredRoles) > 0 {
		hasRequiredRoles = claims.HasAnyRole(config.RequiredRoles...)
	}

	// If no requirements, allow access
	if len(config.RequiredScopes) == 0 && len(config.RequiredRoles) == 0 {
		return true
	}

	return hasRequiredScopes || hasRequiredRoles
}
