package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/oauth"
)

// HTTPClient interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// AuthMiddlewareFunc is a function that wraps an http.HandlerFunc with authentication
type AuthMiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// NewAuthMiddleware creates a new auth middleware with the given configuration
func NewAuthMiddleware(authConfig config.AuthConfig) AuthMiddlewareFunc {
	return NewAuthMiddlewareWithClient(authConfig, &http.Client{})
}

// NewAuthMiddlewareWithClient creates a new auth middleware with the given configuration and HTTP client
func NewAuthMiddlewareWithClient(authConfig config.AuthConfig, client HTTPClient) AuthMiddlewareFunc {
	// Convert HTTPClient to oauth.HTTPClient
	oauthClient := &httpClientAdapter{client}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers for all requests
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Handle CORS preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Extract Bearer token from Authorization header
			token, ok := extractBearerToken(r)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate token via introspection and get claims
			claims, err := oauth.IntrospectToken(token, authConfig, oauthClient)
			if err != nil {
				log.Printf("Token introspection error: %v", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Store claims in request context
			r = oauth.SetAuthClaims(r, claims)

			// Call the next handler with the enriched request
			next(w, r)
		}
	}
}

// httpClientAdapter adapts HTTPClient to oauth.HTTPClient
type httpClientAdapter struct {
	client HTTPClient
}

func (a *httpClientAdapter) Do(req *http.Request) (*http.Response, error) {
	return a.client.Do(req)
}

// extractBearerToken extracts the Bearer token from the Authorization header
func extractBearerToken(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

// JWTAuthMiddlewareFunc is a function that wraps an http.HandlerFunc with JWT authentication
type JWTAuthMiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// NewJWTAuthMiddleware creates a new JWT auth middleware with the given configuration
func NewJWTAuthMiddleware(authConfig config.AuthConfig) JWTAuthMiddlewareFunc {
	return NewJWTAuthMiddlewareWithClient(authConfig, &http.Client{})
}

// NewJWTAuthMiddlewareWithClient creates a new JWT auth middleware with the given configuration and HTTP client
func NewJWTAuthMiddlewareWithClient(authConfig config.AuthConfig, client HTTPClient) JWTAuthMiddlewareFunc {
	// Build JWKS URL from Keycloak configuration
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs",
		authConfig.KeycloakURL, authConfig.RealmName)

	// Convert HTTPClient to oauth.HTTPClient
	oauthClient := &httpClientAdapter{client}

	// Create JWKS cache with 5-minute TTL
	jwksCache := oauth.NewJWKSCache(jwksURL, oauthClient, 5*time.Minute)

	// Build expected issuer
	expectedIssuer := fmt.Sprintf("%s/realms/%s", authConfig.KeycloakURL, authConfig.RealmName)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Extract the token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("No Authorization header found in request to %s", r.URL.Path)
				http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Printf("Invalid Authorization header format")
				http.Error(w, "Unauthorized: Invalid token format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Validate the JWT
			claims, err := oauth.ValidateJWT(tokenString, jwksCache, expectedIssuer)
			if err != nil {
				log.Printf("JWT validation failed: %v", err)
				http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Validate required scope
			if !oauth.HasScope(claims.Scope, authConfig.RequiredScope) {
				log.Printf("Token missing required scope: %s", authConfig.RequiredScope)
				http.Error(w, "Forbidden: Missing required scope", http.StatusForbidden)
				return
			}

			// Call the next handler
			next(w, r)
		}
	}
}
