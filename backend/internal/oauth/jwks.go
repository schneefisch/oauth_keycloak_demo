package oauth

import (
	"context"
	"fmt"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
)

// jwtClaims represents the claims we expect in the JWT (internal use only)
type jwtClaims struct {
	jwt.RegisteredClaims
	Scope    string   `json:"scope"`
	ClientID string   `json:"client_id"`
	Azp      string   `json:"azp"`
	Aud      []string `json:"aud"`
}

// JWKSValidator wraps keyfunc for JWKS-based JWT validation
type JWKSValidator struct {
	keyfunc        keyfunc.Keyfunc
	expectedIssuer string
}

// NewJWKSValidator creates a new validator with automatic JWKS caching
func NewJWKSValidator(ctx context.Context, jwksURL, expectedIssuer string) (*JWKSValidator, error) {
	kf, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS keyfunc: %w", err)
	}
	return &JWKSValidator{keyfunc: kf, expectedIssuer: expectedIssuer}, nil
}

// NewJWKSValidatorFromConfig creates a JWKSValidator from AuthConfig
func NewJWKSValidatorFromConfig(ctx context.Context, authConfig config.AuthConfig) (*JWKSValidator, error) {
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs",
		authConfig.KeycloakURL, authConfig.RealmName)
	expectedIssuer := fmt.Sprintf("%s/realms/%s",
		authConfig.KeycloakURL, authConfig.RealmName)
	return NewJWKSValidator(ctx, jwksURL, expectedIssuer)
}

// ValidateToken validates a JWT and returns AuthClaims
func (v *JWKSValidator) ValidateToken(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, v.keyfunc.Keyfunc,
		jwt.WithIssuer(v.expectedIssuer),
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}),
	)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Convert to AuthClaims
	var scopes []string
	if claims.Scope != "" {
		scopes = strings.Split(claims.Scope, " ")
	}

	return &AuthClaims{
		Subject: claims.Subject,
		Scopes:  scopes,
	}, nil
}
