package oauth

import (
	"context"
	"net/http"
	"slices"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// authClaimsKey is the context key for storing AuthClaims
const authClaimsKey contextKey = "authClaims"

// AuthClaims represents the authenticated user's claims extracted from the token
type AuthClaims struct {
	Subject  string   // sub claim - unique user identifier
	Username string   // preferred_username claim
	Email    string   // email claim
	Scopes   []string // scope claim - space-separated scopes from token
	Roles    []string // realm_access.roles or resource_access roles
}

// TokenIntrospectionResponse represents the response from Keycloak's token introspection endpoint
type TokenIntrospectionResponse struct {
	Active    bool     `json:"active"`
	Scope     string   `json:"scope,omitempty"`
	ClientID  string   `json:"client_id,omitempty"`
	Username  string   `json:"username,omitempty"`
	TokenType string   `json:"token_type,omitempty"`
	Exp       int64    `json:"exp,omitempty"`
	Iat       int64    `json:"iat,omitempty"`
	Nbf       int64    `json:"nbf,omitempty"`
	Sub       string   `json:"sub,omitempty"`
	Aud       []string `json:"aud,omitempty"`
	Iss       string   `json:"iss,omitempty"`
	Jti       string   `json:"jti,omitempty"`
}

// HasScope checks if the claims contain a specific scope
func (c *AuthClaims) HasScope(scope string) bool {
	return slices.Contains(c.Scopes, scope)
}

// HasRole checks if the claims contain a specific role
func (c *AuthClaims) HasRole(role string) bool {
	return slices.Contains(c.Roles, role)
}

// HasAnyScope checks if the claims contain any of the specified scopes
// Returns true if no scopes are required (empty or nil slice)
func (c *AuthClaims) HasAnyScope(scopes ...string) bool {
	if len(scopes) == 0 {
		return true
	}
	return slices.ContainsFunc(scopes, c.HasScope)
}

// HasAnyRole checks if the claims contain any of the specified roles
// Returns true if no roles are required (empty or nil slice)
func (c *AuthClaims) HasAnyRole(roles ...string) bool {
	if len(roles) == 0 {
		return true
	}
	return slices.ContainsFunc(roles, c.HasRole)
}

// HasAllScopes checks if the claims contain all of the specified scopes
// Returns true if no scopes are required (empty or nil slice)
func (c *AuthClaims) HasAllScopes(scopes ...string) bool {
	if len(scopes) == 0 {
		return true
	}
	for _, scope := range scopes {
		if !c.HasScope(scope) {
			return false
		}
	}
	return true
}

// HasAllRoles checks if the claims contain all of the specified roles
// Returns true if no roles are required (empty or nil slice)
func (c *AuthClaims) HasAllRoles(roles ...string) bool {
	if len(roles) == 0 {
		return true
	}
	for _, role := range roles {
		if !c.HasRole(role) {
			return false
		}
	}
	return true
}

// GetAuthClaims retrieves AuthClaims from the request context
// Returns nil if no claims are present or if the value is not of type *AuthClaims
func GetAuthClaims(r *http.Request) *AuthClaims {
	claims, ok := r.Context().Value(authClaimsKey).(*AuthClaims)
	if !ok {
		return nil
	}
	return claims
}

// SetAuthClaims stores AuthClaims in the request context and returns a new request
// with the enriched context
func SetAuthClaims(r *http.Request, claims *AuthClaims) *http.Request {
	ctx := context.WithValue(r.Context(), authClaimsKey, claims)
	return r.WithContext(ctx)
}
