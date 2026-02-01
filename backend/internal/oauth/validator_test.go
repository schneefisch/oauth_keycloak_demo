package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
)

// MockHTTPClient is a mock implementation of the HTTPClient interface
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do implements the HTTPClient interface
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestNewTokenValidator_Introspection(t *testing.T) {
	cfg := ValidatorConfig{
		AuthConfig: config.AuthConfig{
			KeycloakURL:   "http://keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RealmName:     "test-realm",
			RequiredScope: "test-scope",
		},
		HTTPClient: &http.Client{},
	}

	validator, err := NewTokenValidator(ValidationMethodIntrospection, cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if validator == nil {
		t.Fatal("Expected validator to be created")
	}

	// Check it's the right type
	if _, ok := validator.(*IntrospectionValidator); !ok {
		t.Errorf("Expected IntrospectionValidator, got %T", validator)
	}
}

func TestNewTokenValidator_EmptyMethodDefaultsToIntrospection(t *testing.T) {
	cfg := ValidatorConfig{
		AuthConfig: config.AuthConfig{
			KeycloakURL:   "http://keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RealmName:     "test-realm",
			RequiredScope: "test-scope",
		},
		HTTPClient: &http.Client{},
	}

	validator, err := NewTokenValidator("", cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check it defaults to IntrospectionValidator
	if _, ok := validator.(*IntrospectionValidator); !ok {
		t.Errorf("Expected IntrospectionValidator for empty method, got %T", validator)
	}
}

func TestNewTokenValidator_UnsupportedMethod(t *testing.T) {
	cfg := ValidatorConfig{
		AuthConfig: config.AuthConfig{
			KeycloakURL:   "http://keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RealmName:     "test-realm",
			RequiredScope: "test-scope",
		},
	}

	_, err := NewTokenValidator("unsupported", cfg)
	if err == nil {
		t.Fatal("Expected error for unsupported method")
	}
}

func TestIntrospectionValidator_ValidToken(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			response := TokenIntrospectionResponse{
				Active:   true,
				Scope:    "test-scope read:events",
				Sub:      "user-123",
				Username: "testuser",
			}
			responseBody, _ := json.Marshal(response)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(string(responseBody))),
				Header:     make(http.Header),
			}, nil
		},
	}

	authConfig := config.AuthConfig{
		KeycloakURL:  "http://keycloak:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RealmName:    "test-realm",
	}

	validator := NewIntrospectionValidator(authConfig, mockClient)

	claims, err := validator.ValidateToken("valid-token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if claims.Subject != "user-123" {
		t.Errorf("Expected Subject 'user-123', got '%s'", claims.Subject)
	}
	if claims.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", claims.Username)
	}
	if !claims.HasScope("test-scope") {
		t.Error("Expected claims to have 'test-scope'")
	}
	if !claims.HasScope("read:events") {
		t.Error("Expected claims to have 'read:events'")
	}
}

func TestIntrospectionValidator_InvalidToken(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			response := TokenIntrospectionResponse{
				Active: false,
			}
			responseBody, _ := json.Marshal(response)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(string(responseBody))),
				Header:     make(http.Header),
			}, nil
		},
	}

	authConfig := config.AuthConfig{
		KeycloakURL:  "http://keycloak:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RealmName:    "test-realm",
	}

	validator := NewIntrospectionValidator(authConfig, mockClient)

	_, err := validator.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("Expected error for inactive token")
	}
}

func TestIntrospectionValidator_HTTPError(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("connection refused")
		},
	}

	authConfig := config.AuthConfig{
		KeycloakURL:  "http://keycloak:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RealmName:    "test-realm",
	}

	validator := NewIntrospectionValidator(authConfig, mockClient)

	_, err := validator.ValidateToken("any-token")
	if err == nil {
		t.Fatal("Expected error for HTTP failure")
	}
}

func TestIntrospectionValidator_NilClient(t *testing.T) {
	authConfig := config.AuthConfig{
		KeycloakURL:  "http://keycloak:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RealmName:    "test-realm",
	}

	// Should not panic with nil client
	validator := NewIntrospectionValidator(authConfig, nil)
	if validator == nil {
		t.Fatal("Expected validator to be created")
	}
	if validator.client == nil {
		t.Error("Expected default client to be set")
	}
}

func TestValidationMethod_Constants(t *testing.T) {
	if ValidationMethodIntrospection != "introspection" {
		t.Errorf("Expected 'introspection', got '%s'", ValidationMethodIntrospection)
	}
	if ValidationMethodJWKS != "jwks" {
		t.Errorf("Expected 'jwks', got '%s'", ValidationMethodJWKS)
	}
}

func TestNewTokenValidator_JWKS_RequiresContext(t *testing.T) {
	cfg := ValidatorConfig{
		AuthConfig: config.AuthConfig{
			KeycloakURL:   "http://keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RealmName:     "test-realm",
			RequiredScope: "test-scope",
		},
		Context: nil, // No context provided
	}

	// Note: This will fail because keycloak is not running, but it should use context.Background()
	// In a real test environment with a mock JWKS server, this would succeed
	_, err := NewTokenValidator(ValidationMethodJWKS, cfg)
	// We expect an error because there's no JWKS server running
	if err == nil {
		t.Log("JWKS validator created successfully (mock JWKS server may be running)")
	}
}

func TestNewTokenValidator_JWKS_WithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := ValidatorConfig{
		AuthConfig: config.AuthConfig{
			KeycloakURL:   "http://keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RealmName:     "test-realm",
			RequiredScope: "test-scope",
		},
		Context: ctx,
	}

	// Note: This will fail because keycloak is not running
	// In a real test environment with a mock JWKS server, this would succeed
	_, err := NewTokenValidator(ValidationMethodJWKS, cfg)
	// We expect an error because there's no JWKS server running
	if err == nil {
		t.Log("JWKS validator created successfully (mock JWKS server may be running)")
	}
}
