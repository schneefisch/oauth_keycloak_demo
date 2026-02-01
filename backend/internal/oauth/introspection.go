package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
)

// HTTPClient interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// IntrospectToken validates the token against Keycloak's introspection endpoint
// and returns the extracted claims
func IntrospectToken(token string, authConfig config.AuthConfig, client HTTPClient) (*AuthClaims, error) {
	introspectionURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token/introspect",
		authConfig.KeycloakURL, authConfig.RealmName)

	// Prepare the introspection request
	data := url.Values{}
	data.Set("token", token)
	data.Set("client_id", authConfig.ClientID)
	data.Set("client_secret", authConfig.ClientSecret)

	req, err := http.NewRequest(http.MethodPost, introspectionURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create introspection request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Calling the introspection endpoint
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("introspection request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspection returned status %d", resp.StatusCode)
	}

	// Extract and parse the response-body into TokenIntrospectionResponse
	var introspectionResp TokenIntrospectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&introspectionResp); err != nil {
		return nil, fmt.Errorf("failed to decode introspection response: %w", err)
	}

	// Check if the token is active, if not, return error
	if !introspectionResp.Active {
		return nil, fmt.Errorf("token is not active")
	}

	// Parse scopes from space-separated string
	var scopes []string
	if introspectionResp.Scope != "" {
		scopes = strings.Split(introspectionResp.Scope, " ")
	}

	claims := &AuthClaims{
		Subject:  introspectionResp.Sub,
		Username: introspectionResp.Username,
		Scopes:   scopes,
		// Roles would typically come from realm_access.roles or resource_access in JWT
		// For introspection, we'd need additional claims in the response
	}

	return claims, nil
}
