package oauth

import (
	"context"
	"fmt"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
)

// ValidationMethod defines the method used for token validation
type ValidationMethod string

const (
	// ValidationMethodIntrospection validates tokens via Keycloak's introspection endpoint
	ValidationMethodIntrospection ValidationMethod = "introspection"
	// ValidationMethodJWKS validates tokens locally using JWKS
	ValidationMethodJWKS ValidationMethod = "jwks"
)

// TokenValidator is the interface for validating OAuth tokens
type TokenValidator interface {
	ValidateToken(token string) (*AuthClaims, error)
}

// ValidatorConfig holds configuration for creating a TokenValidator
type ValidatorConfig struct {
	AuthConfig config.AuthConfig
	HTTPClient HTTPClient
	Context    context.Context
}

// NewTokenValidator creates a TokenValidator based on the specified method
func NewTokenValidator(method ValidationMethod, cfg ValidatorConfig) (TokenValidator, error) {
	switch method {
	case ValidationMethodIntrospection, "":
		return NewIntrospectionValidator(cfg.AuthConfig, cfg.HTTPClient), nil
	case ValidationMethodJWKS:
		if cfg.Context == nil {
			cfg.Context = context.Background()
		}
		return NewJWKSValidatorFromConfig(cfg.Context, cfg.AuthConfig)
	default:
		return nil, fmt.Errorf("unsupported validation method: %s", method)
	}
}
