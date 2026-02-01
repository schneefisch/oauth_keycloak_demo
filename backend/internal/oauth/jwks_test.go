package oauth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
)

// testKeyPair holds an RSA key pair for testing
type testKeyPair struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
}

// generateTestKeyPair creates a new RSA key pair for testing
func generateTestKeyPair(t *testing.T) *testKeyPair {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	return &testKeyPair{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		keyID:      "test-key-id",
	}
}

// jwksResponse creates a JWKS JSON response from the test key pair
func (kp *testKeyPair) jwksResponse() []byte {
	// Encode the public key components
	nBytes := kp.publicKey.N.Bytes()
	eBytes := big.NewInt(int64(kp.publicKey.E)).Bytes()

	jwks := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"use": "sig",
				"kid": kp.keyID,
				"alg": "RS256",
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			},
		},
	}

	data, _ := json.Marshal(jwks)
	return data
}

// createTestToken creates a signed JWT with the given claims
func (kp *testKeyPair) createTestToken(t *testing.T, claims jwt.MapClaims, method jwt.SigningMethod) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = kp.keyID

	tokenString, err := token.SignedString(kp.privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenString
}

// createMockJWKSServer creates a test server that serves JWKS
func createMockJWKSServer(t *testing.T, keyPair *testKeyPair) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(keyPair.jwksResponse())
	}))
}

func TestNewJWKSValidator_Success(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	ctx := context.Background()
	validator, err := NewJWKSValidator(ctx, server.URL, "https://test-issuer.com")

	if err != nil {
		t.Fatalf("NewJWKSValidator() error = %v, want nil", err)
	}
	if validator == nil {
		t.Fatal("NewJWKSValidator() returned nil validator")
	}
	if validator.expectedIssuer != "https://test-issuer.com" {
		t.Errorf("expectedIssuer = %v, want %v", validator.expectedIssuer, "https://test-issuer.com")
	}
}

func TestNewJWKSValidator_InvalidURL_ValidateTokenFails(t *testing.T) {
	ctx := context.Background()

	// The keyfunc library uses background refresh and doesn't fail on creation
	// when the URL is unreachable. Instead, validation will fail when no keys
	// are available.
	validator, err := NewJWKSValidator(ctx, "http://localhost:1/nonexistent", "https://test-issuer.com")

	// Creation succeeds even with invalid URL (background refresh pattern)
	if err != nil {
		t.Skipf("Validator creation failed (expected on some systems): %v", err)
	}

	// But token validation should fail because no keys are available
	_, err = validator.ValidateToken("any.token.here")
	if err == nil {
		t.Fatal("ValidateToken() expected error when no JWKS keys available, got nil")
	}
}

func TestNewJWKSValidatorFromConfig_URLConstruction(t *testing.T) {
	keyPair := generateTestKeyPair(t)

	// Create a server that verifies the expected path
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write(keyPair.jwksResponse())
	}))
	defer server.Close()

	authConfig := config.AuthConfig{
		KeycloakURL: server.URL,
		RealmName:   "test-realm",
	}

	ctx := context.Background()
	validator, err := NewJWKSValidatorFromConfig(ctx, authConfig)

	if err != nil {
		t.Fatalf("NewJWKSValidatorFromConfig() error = %v, want nil", err)
	}

	// Verify the JWKS URL was constructed correctly
	expectedPath := "/realms/test-realm/protocol/openid-connect/certs"
	if requestedPath != expectedPath {
		t.Errorf("JWKS path = %v, want %v", requestedPath, expectedPath)
	}

	// Verify the expected issuer was constructed correctly
	expectedIssuer := fmt.Sprintf("%s/realms/test-realm", server.URL)
	if validator.expectedIssuer != expectedIssuer {
		t.Errorf("expectedIssuer = %v, want %v", validator.expectedIssuer, expectedIssuer)
	}
}

func TestJWKSValidator_ValidToken(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	ctx := context.Background()
	validator, err := NewJWKSValidator(ctx, server.URL, server.URL)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	claims := jwt.MapClaims{
		"sub":   "user-123",
		"iss":   server.URL,
		"scope": "events:read events:write",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := keyPair.createTestToken(t, claims, jwt.SigningMethodRS256)

	authClaims, err := validator.ValidateToken(token)

	if err != nil {
		t.Fatalf("ValidateToken() error = %v, want nil", err)
	}
	if authClaims == nil {
		t.Fatal("ValidateToken() returned nil AuthClaims")
	}
	if authClaims.Subject != "user-123" {
		t.Errorf("Subject = %v, want %v", authClaims.Subject, "user-123")
	}
	if len(authClaims.Scopes) != 2 {
		t.Errorf("Scopes length = %v, want 2", len(authClaims.Scopes))
	}
	if authClaims.Scopes[0] != "events:read" {
		t.Errorf("Scopes[0] = %v, want %v", authClaims.Scopes[0], "events:read")
	}
	if authClaims.Scopes[1] != "events:write" {
		t.Errorf("Scopes[1] = %v, want %v", authClaims.Scopes[1], "events:write")
	}
}

func TestJWKSValidator_InvalidSignature(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	wrongKeyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	ctx := context.Background()
	validator, err := NewJWKSValidator(ctx, server.URL, server.URL)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	claims := jwt.MapClaims{
		"sub": "user-123",
		"iss": server.URL,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	// Sign with wrong key
	token := wrongKeyPair.createTestToken(t, claims, jwt.SigningMethodRS256)

	_, err = validator.ValidateToken(token)

	if err == nil {
		t.Fatal("ValidateToken() expected error for invalid signature, got nil")
	}
}

func TestJWKSValidator_ExpiredToken(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	ctx := context.Background()
	validator, err := NewJWKSValidator(ctx, server.URL, server.URL)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	claims := jwt.MapClaims{
		"sub": "user-123",
		"iss": server.URL,
		"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := keyPair.createTestToken(t, claims, jwt.SigningMethodRS256)

	_, err = validator.ValidateToken(token)

	if err == nil {
		t.Fatal("ValidateToken() expected error for expired token, got nil")
	}
}

func TestJWKSValidator_WrongIssuer(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	ctx := context.Background()
	validator, err := NewJWKSValidator(ctx, server.URL, "https://expected-issuer.com")
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	claims := jwt.MapClaims{
		"sub": "user-123",
		"iss": "https://wrong-issuer.com", // Wrong issuer
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := keyPair.createTestToken(t, claims, jwt.SigningMethodRS256)

	_, err = validator.ValidateToken(token)

	if err == nil {
		t.Fatal("ValidateToken() expected error for wrong issuer, got nil")
	}
}

func TestJWKSValidator_UnsupportedAlgorithm(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	ctx := context.Background()
	validator, err := NewJWKSValidator(ctx, server.URL, server.URL)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// Create a token with HS256 (HMAC) instead of RS256
	claims := jwt.MapClaims{
		"sub": "user-123",
		"iss": server.URL,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("secret-key"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = validator.ValidateToken(tokenString)

	if err == nil {
		t.Fatal("ValidateToken() expected error for unsupported algorithm, got nil")
	}
}

func TestJWKSValidator_MalformedToken(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	ctx := context.Background()
	validator, err := NewJWKSValidator(ctx, server.URL, server.URL)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	tests := []struct {
		name  string
		token string
	}{
		{"empty string", ""},
		{"random string", "not-a-jwt-token"},
		{"invalid base64", "invalid.base64.token"},
		{"only two parts", "header.payload"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.ValidateToken(tt.token)

			if err == nil {
				t.Errorf("ValidateToken(%q) expected error, got nil", tt.token)
			}
		})
	}
}

func TestJWKSValidator_ScopeParsing(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	ctx := context.Background()
	validator, err := NewJWKSValidator(ctx, server.URL, server.URL)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	tests := []struct {
		name       string
		scope      string
		wantScopes []string
	}{
		{
			name:       "single scope",
			scope:      "events:read",
			wantScopes: []string{"events:read"},
		},
		{
			name:       "multiple scopes",
			scope:      "events:read events:write profile email",
			wantScopes: []string{"events:read", "events:write", "profile", "email"},
		},
		{
			name:       "empty scope",
			scope:      "",
			wantScopes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := jwt.MapClaims{
				"sub":   "user-123",
				"iss":   server.URL,
				"scope": tt.scope,
				"exp":   time.Now().Add(time.Hour).Unix(),
				"iat":   time.Now().Unix(),
			}

			token := keyPair.createTestToken(t, claims, jwt.SigningMethodRS256)

			authClaims, err := validator.ValidateToken(token)

			if err != nil {
				t.Fatalf("ValidateToken() error = %v, want nil", err)
			}

			if len(authClaims.Scopes) != len(tt.wantScopes) {
				t.Errorf("Scopes length = %v, want %v", len(authClaims.Scopes), len(tt.wantScopes))
				return
			}

			for i, scope := range authClaims.Scopes {
				if scope != tt.wantScopes[i] {
					t.Errorf("Scopes[%d] = %v, want %v", i, scope, tt.wantScopes[i])
				}
			}
		})
	}
}

func TestJWKSValidator_RS384Algorithm(t *testing.T) {
	keyPair := generateTestKeyPair(t)

	// Create server with RS384 key
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nBytes := keyPair.publicKey.N.Bytes()
		eBytes := big.NewInt(int64(keyPair.publicKey.E)).Bytes()

		jwks := map[string]any{
			"keys": []map[string]any{
				{
					"kty": "RSA",
					"use": "sig",
					"kid": keyPair.keyID,
					"alg": "RS384",
					"n":   base64.RawURLEncoding.EncodeToString(nBytes),
					"e":   base64.RawURLEncoding.EncodeToString(eBytes),
				},
			},
		}
		data, _ := json.Marshal(jwks)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	ctx := context.Background()
	validator, err := NewJWKSValidator(ctx, server.URL, server.URL)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	claims := jwt.MapClaims{
		"sub": "user-123",
		"iss": server.URL,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := keyPair.createTestToken(t, claims, jwt.SigningMethodRS384)

	authClaims, err := validator.ValidateToken(token)

	if err != nil {
		t.Fatalf("ValidateToken() with RS384 error = %v, want nil", err)
	}
	if authClaims.Subject != "user-123" {
		t.Errorf("Subject = %v, want %v", authClaims.Subject, "user-123")
	}
}

func TestJWKSValidator_RS512Algorithm(t *testing.T) {
	keyPair := generateTestKeyPair(t)

	// Create server with RS512 key
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nBytes := keyPair.publicKey.N.Bytes()
		eBytes := big.NewInt(int64(keyPair.publicKey.E)).Bytes()

		jwks := map[string]any{
			"keys": []map[string]any{
				{
					"kty": "RSA",
					"use": "sig",
					"kid": keyPair.keyID,
					"alg": "RS512",
					"n":   base64.RawURLEncoding.EncodeToString(nBytes),
					"e":   base64.RawURLEncoding.EncodeToString(eBytes),
				},
			},
		}
		data, _ := json.Marshal(jwks)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	ctx := context.Background()
	validator, err := NewJWKSValidator(ctx, server.URL, server.URL)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	claims := jwt.MapClaims{
		"sub": "user-123",
		"iss": server.URL,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := keyPair.createTestToken(t, claims, jwt.SigningMethodRS512)

	authClaims, err := validator.ValidateToken(token)

	if err != nil {
		t.Fatalf("ValidateToken() with RS512 error = %v, want nil", err)
	}
	if authClaims.Subject != "user-123" {
		t.Errorf("Subject = %v, want %v", authClaims.Subject, "user-123")
	}
}

func TestJwtClaims_StructTags(t *testing.T) {
	// Verify jwtClaims struct fields have correct JSON tags
	// This is tested by creating a claims object and marshaling/unmarshaling

	claims := &jwtClaims{
		Scope:    "read write",
		ClientID: "my-client",
		Azp:      "authorized-party",
		Aud:      []string{"aud1", "aud2"},
	}

	data, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("failed to marshal jwtClaims: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	// Check the JSON field names
	if _, ok := unmarshaled["scope"]; !ok {
		t.Error("expected 'scope' field in JSON")
	}
	if _, ok := unmarshaled["client_id"]; !ok {
		t.Error("expected 'client_id' field in JSON")
	}
	if _, ok := unmarshaled["azp"]; !ok {
		t.Error("expected 'azp' field in JSON")
	}
	if _, ok := unmarshaled["aud"]; !ok {
		t.Error("expected 'aud' field in JSON")
	}
}
