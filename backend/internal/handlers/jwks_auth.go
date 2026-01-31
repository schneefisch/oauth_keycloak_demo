package handlers

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
)

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a single JSON Web Key
type JWK struct {
	Kty string `json:"kty"` // Key Type (e.g., "RSA")
	Use string `json:"use"` // Key Use (e.g., "sig" for signature)
	Kid string `json:"kid"` // Key ID
	Alg string `json:"alg"` // Algorithm (e.g., "RS256")
	N   string `json:"n"`   // RSA modulus
	E   string `json:"e"`   // RSA public exponent
}

// JWKSCache caches the JWKS to avoid fetching on every request
type JWKSCache struct {
	keys       map[string]*rsa.PublicKey
	mu         sync.RWMutex
	lastFetch  time.Time
	cacheTTL   time.Duration
	jwksURL    string
	httpClient HTTPClient
}

// NewJWKSCache creates a new JWKS cache
func NewJWKSCache(jwksURL string, client HTTPClient, cacheTTL time.Duration) *JWKSCache {
	return &JWKSCache{
		keys:       make(map[string]*rsa.PublicKey),
		cacheTTL:   cacheTTL,
		jwksURL:    jwksURL,
		httpClient: client,
	}
}

// GetKey retrieves a public key by key ID, fetching from JWKS if needed
func (c *JWKSCache) GetKey(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	if key, ok := c.keys[kid]; ok && time.Since(c.lastFetch) < c.cacheTTL {
		c.mu.RUnlock()
		return key, nil
	}
	c.mu.RUnlock()

	// Cache miss or expired - fetch JWKS
	if err := c.refresh(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	if key, ok := c.keys[kid]; ok {
		return key, nil
	}

	return nil, fmt.Errorf("key with kid %s not found in JWKS", kid)
}

// refresh fetches the JWKS and updates the cache
func (c *JWKSCache) refresh() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(c.lastFetch) < c.cacheTTL {
		return nil
	}

	req, err := http.NewRequest("GET", c.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("error creating JWKS request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading JWKS response: %w", err)
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("error parsing JWKS: %w", err)
	}

	// Convert JWKs to RSA public keys
	newKeys := make(map[string]*rsa.PublicKey)
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" || jwk.Use != "sig" {
			continue
		}

		pubKey, err := jwkToRSAPublicKey(jwk)
		if err != nil {
			log.Printf("Warning: failed to parse JWK %s: %v", jwk.Kid, err)
			continue
		}
		newKeys[jwk.Kid] = pubKey
	}

	c.keys = newKeys
	c.lastFetch = time.Now()
	log.Printf("JWKS cache refreshed, %d keys loaded", len(newKeys))

	return nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode the modulus (n)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode the exponent (e)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert exponent bytes to int
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}, nil
}

// JWTClaims represents the claims we expect in the JWT
type JWTClaims struct {
	jwt.RegisteredClaims
	Scope    string   `json:"scope"`
	ClientID string   `json:"client_id"`
	Azp      string   `json:"azp"`
	Aud      []string `json:"aud"`
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

	// Create JWKS cache with 5-minute TTL
	jwksCache := NewJWKSCache(jwksURL, client, 5*time.Minute)

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
			claims, err := validateJWT(tokenString, jwksCache, expectedIssuer)
			if err != nil {
				log.Printf("JWT validation failed: %v", err)
				http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Validate required scope
			if !hasScope(claims.Scope, authConfig.RequiredScope) {
				log.Printf("Token missing required scope: %s", authConfig.RequiredScope)
				http.Error(w, "Forbidden: Missing required scope", http.StatusForbidden)
				return
			}

			// Call the next handler
			next(w, r)
		}
	}
}

// validateJWT validates the JWT signature and standard claims
func validateJWT(tokenString string, jwksCache *JWKSCache, expectedIssuer string) (*JWTClaims, error) {
	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("token missing kid header")
		}

		// Fetch the public key from JWKS
		return jwksCache.GetKey(kid)
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims")
	}

	// Validate issuer
	if claims.Issuer != expectedIssuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", expectedIssuer, claims.Issuer)
	}

	return claims, nil
}

// hasScope checks if the token's scope string contains the required scope
func hasScope(scopeString, requiredScope string) bool {
	scopes := strings.Split(scopeString, " ")
	for _, scope := range scopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}
