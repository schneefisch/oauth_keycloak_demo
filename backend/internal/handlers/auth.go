package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
)

// TokenIntrospectionResponse represents the response from Keycloak's token introspection endpoint
type TokenIntrospectionResponse struct {
	Active    bool     `json:"active"`
	Scope     string   `json:"scope"`
	ClientID  string   `json:"client_id"`
	Username  string   `json:"username"`
	TokenType string   `json:"token_type"`
	Exp       int64    `json:"exp"`
	Iat       int64    `json:"iat"`
	Nbf       int64    `json:"nbf"`
	Sub       string   `json:"sub"`
	Aud       []string `json:"aud"`
	Iss       string   `json:"iss"`
	Jti       string   `json:"jti"`
}

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
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers to allow requests from the frontend
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

			// The Authorization header typically has the format "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Printf("Invalid Authorization header format: %s", authHeader)
				http.Error(w, "Unauthorized: Invalid token format", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			log.Printf("Received token: %s", token)

			// Validate the token with Keycloak's introspection endpoint
			valid, err := validateToken(token, authConfig, client)
			if err != nil {
				log.Printf("Error validating token: %v", err)
				http.Error(w, "Unauthorized: Token validation failed", http.StatusUnauthorized)
				return
			}

			if !valid {
				log.Printf("Invalid token or missing required scope")
				http.Error(w, "Unauthorized: Invalid token or missing required scope", http.StatusUnauthorized)
				return
			}

			// Call the next handler
			next(w, r)
		}
	}
}

// validateToken validates the token against Keycloak's introspection endpoint
// and checks if it contains the required scope
func validateToken(token string, authConfig config.AuthConfig, client HTTPClient) (bool, error) {
	// Check if client secret is set
	if authConfig.ClientSecret == "" {
		return false, fmt.Errorf("client secret is not set in configuration")
	}

	// Construct the introspection endpoint URL
	introspectionURL := fmt.Sprintf("%s/realms/events/protocol/openid-connect/token/introspect", authConfig.KeycloakURL)

	// Prepare the request data
	data := url.Values{}
	data.Set("token", token)
	data.Set("client_id", authConfig.ClientID)
	data.Set("client_secret", authConfig.ClientSecret)

	// Create the HTTP request
	req, err := http.NewRequest("POST", introspectionURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("error reading response: %w", err)
	}

	// Parse the response
	var introspectionResp TokenIntrospectionResponse
	if err := json.Unmarshal(body, &introspectionResp); err != nil {
		return false, fmt.Errorf("error parsing response: %w", err)
	}

	// Check if the token is active
	if !introspectionResp.Active {
		return false, nil
	}

	// Check if the token has the required scope
	// First, log the scopes for debugging
	log.Printf("Token scopes: %s, Required scope: %s", introspectionResp.Scope, authConfig.RequiredScope)

	// Check if the required scope is contained within the token's scope string
	// This is more flexible than requiring an exact match
	hasRequiredScope := strings.Contains(introspectionResp.Scope, authConfig.RequiredScope)

	// If the above check fails, try the original method of splitting by space
	if !hasRequiredScope {
		scopes := strings.Split(introspectionResp.Scope, " ")
		for _, scope := range scopes {
			if scope == authConfig.RequiredScope {
				hasRequiredScope = true
				break
			}
		}
	}

	return hasRequiredScope, nil
}
