package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/oauth"
)

// NewAuthMiddleware creates a new auth middleware with the given configuration
func NewAuthMiddleware(authConfig config.AuthConfig) func(http.Handler) http.Handler {
	return NewAuthMiddlewareWithClient(authConfig, &http.Client{})
}

// NewAuthMiddlewareWithClient creates a new auth middleware with the given configuration and HTTP client
func NewAuthMiddlewareWithClient(authConfig config.AuthConfig, client oauth.HTTPClient) func(http.Handler) http.Handler {
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

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

// NewJWTAuthMiddleware creates a new JWT auth middleware with the given configuration
func NewJWTAuthMiddleware(authConfig config.AuthConfig) (func(http.Handler) http.Handler, error) {
	return NewJWTAuthMiddlewareWithContext(context.Background(), authConfig)
}

// NewJWTAuthMiddlewareWithContext creates a new JWT auth middleware with the given configuration and context
func NewJWTAuthMiddlewareWithContext(ctx context.Context, authConfig config.AuthConfig) (func(http.Handler) http.Handler, error) {
	// Build JWKS URL from Keycloak configuration
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs",
		authConfig.KeycloakURL, authConfig.RealmName)

	// Build expected issuer
	expectedIssuer := fmt.Sprintf("%s/realms/%s", authConfig.KeycloakURL, authConfig.RealmName)

	// Create JWKS validator using keyfunc library
	validator, err := oauth.NewJWKSValidator(ctx, jwksURL, expectedIssuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS validator: %w", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the token from the Authorization header
			token, ok := extractBearerToken(r)
			if !ok {
				log.Printf("No valid Authorization header found in request to %s", r.URL.Path)
				http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
				return
			}

			// Validate the JWT
			claims, err := validator.ValidateToken(token)
			if err != nil {
				log.Printf("JWT validation failed: %v", err)
				http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Convert JWTClaims to AuthClaims and store in context
			authClaims := &oauth.AuthClaims{
				Subject:  claims.Subject,
				Username: claims.Azp,
				Scopes:   strings.Split(claims.Scope, " "),
			}
			r = oauth.SetAuthClaims(r, authClaims)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}, nil
}
