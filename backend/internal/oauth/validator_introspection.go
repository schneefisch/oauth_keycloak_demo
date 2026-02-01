package oauth

import (
	"net/http"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
)

// IntrospectionValidator validates tokens using Keycloak's introspection endpoint
type IntrospectionValidator struct {
	authConfig config.AuthConfig
	client     HTTPClient
}

// NewIntrospectionValidator creates a new IntrospectionValidator
func NewIntrospectionValidator(authConfig config.AuthConfig, client HTTPClient) *IntrospectionValidator {
	if client == nil {
		client = &http.Client{}
	}
	return &IntrospectionValidator{
		authConfig: authConfig,
		client:     client,
	}
}

// ValidateToken validates the token via introspection and returns AuthClaims
func (v *IntrospectionValidator) ValidateToken(token string) (*AuthClaims, error) {
	return IntrospectToken(token, v.authConfig, v.client)
}
