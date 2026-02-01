package oauth

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims we expect in the JWT
type JWTClaims struct {
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

// ValidateToken validates a JWT and returns the claims
func (v *JWKSValidator) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, v.keyfunc.Keyfunc,
		jwt.WithIssuer(v.expectedIssuer),
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}),
	)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// HasScope checks if the token's scope string contains the required scope
func HasScope(scopeString, requiredScope string) bool {
	scopes := strings.Split(scopeString, " ")
	return slices.Contains(scopes, requiredScope)
}
