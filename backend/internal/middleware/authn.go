package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/oauth"
)

// NewAuthMiddleware creates a new auth middleware with the given configuration
func NewAuthMiddleware(authConfig config.AuthConfig) func(http.Handler) http.Handler {
	return NewIntrospectionAuthMiddlewareWithClient(authConfig, &http.Client{})
}

// NewAuthMiddlewareWithValidator creates a new auth middleware using a TokenValidator
func NewAuthMiddlewareWithValidator(validator oauth.TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Bearer token from Authorization header
			token, ok := extractBearerToken(r)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate token using the validator and get claims
			claims, err := validator.ValidateToken(token)
			if err != nil {
				log.Printf("Token validation error: %v", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Store claims in request context
			r = oauth.SetAuthClaims(r, claims)

			// Call the next handler with the enriched request
			next.ServeHTTP(w, r)
		})
	}
}

// NewIntrospectionAuthMiddlewareWithClient creates a new auth middleware with the given configuration and HTTP client
func NewIntrospectionAuthMiddlewareWithClient(authConfig config.AuthConfig, client oauth.HTTPClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Bearer token from Authorization header
			token, ok := extractBearerToken(r)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate token via introspection and get claims
			claims, err := oauth.IntrospectToken(token, authConfig, client)
			if err != nil {
				log.Printf("Token introspection error: %v", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Store claims in request context
			r = oauth.SetAuthClaims(r, claims)

			// Call the next handler with the enriched request
			next.ServeHTTP(w, r)
		})
	}
}

// extractBearerToken extracts the Bearer token from the Authorization header
func extractBearerToken(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", false
	}

	// Must follow the pattern "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return "", false
	}

	// returning only the Token-Part
	return parts[1], true
}
