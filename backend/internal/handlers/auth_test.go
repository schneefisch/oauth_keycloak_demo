package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

// createMockResponse creates a mock HTTP response with the given status code and body
func createMockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}

// createValidTokenResponse creates a mock response for a valid token
func createValidTokenResponse(requiredScope string) *http.Response {
	response := TokenIntrospectionResponse{
		Active:    true,
		Scope:     requiredScope,
		ClientID:  "test-client",
		Username:  "test-user",
		TokenType: "Bearer",
		Exp:       1735689600, // Some future timestamp
		Iat:       1619712000, // Some past timestamp
		Nbf:       1619712000, // Some past timestamp
		Sub:       "test-subject",
		Aud:       []string{"test-audience"},
		Iss:       "test-issuer",
		Jti:       "test-jti",
	}

	responseBody, _ := json.Marshal(response)
	return createMockResponse(http.StatusOK, string(responseBody))
}

// createInvalidTokenResponse creates a mock response for an invalid token
func createInvalidTokenResponse() *http.Response {
	response := TokenIntrospectionResponse{
		Active: false,
	}

	responseBody, _ := json.Marshal(response)
	return createMockResponse(http.StatusOK, string(responseBody))
}

// createMissingRequiredScopeResponse creates a mock response for a token missing the required scope
func createMissingRequiredScopeResponse() *http.Response {
	response := TokenIntrospectionResponse{
		Active:    true,
		Scope:     "other-scope",
		ClientID:  "test-client",
		Username:  "test-user",
		TokenType: "Bearer",
		Exp:       1735689600, // Some future timestamp
		Iat:       1619712000, // Some past timestamp
		Nbf:       1619712000, // Some past timestamp
		Sub:       "test-subject",
		Aud:       []string{"test-audience"},
		Iss:       "test-issuer",
		Jti:       "test-jti",
	}

	responseBody, _ := json.Marshal(response)
	return createMockResponse(http.StatusOK, string(responseBody))
}

// createErrorResponse creates a mock response for an error
func createErrorResponse() *http.Response {
	return createMockResponse(http.StatusInternalServerError, "Internal Server Error")
}

func TestAuthMiddlewareWithValidToken(t *testing.T) {
	// Create a test configuration with mock values
	testConfig := config.TestConfig(&config.Config{
		Auth: config.AuthConfig{
			KeycloakURL:   "http://mock-keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RequiredScope: "test-scope",
		},
	})

	// Create a mock handler that will be wrapped by the auth middleware
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}

	// Create a mock HTTP client that returns a valid token response
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return createValidTokenResponse("test-scope"), nil
		},
	}

	// Create the auth middleware with the test configuration and mock client
	authMiddleware := NewAuthMiddlewareWithClient(testConfig.Auth, mockClient)

	// Create a test request with a valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the middleware with the mock handler
	authMiddleware(mockHandler)(rr, req)

	// Check the response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Check the response body
	if rr.Body.String() != "Success" {
		t.Errorf("Expected response body %q, got %q", "Success", rr.Body.String())
	}
}

func TestAuthMiddlewareWithInvalidToken(t *testing.T) {
	// Create a test configuration with mock values
	testConfig := config.TestConfig(&config.Config{
		Auth: config.AuthConfig{
			KeycloakURL:   "http://mock-keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RequiredScope: "test-scope",
		},
	})

	// Create a mock handler that will be wrapped by the auth middleware
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}

	// Create a mock HTTP client that returns an invalid token response
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return createInvalidTokenResponse(), nil
		},
	}

	// Create the auth middleware with the test configuration and mock client
	authMiddleware := NewAuthMiddlewareWithClient(testConfig.Auth, mockClient)

	// Create a test request with an invalid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the middleware with the mock handler
	authMiddleware(mockHandler)(rr, req)

	// Check the response
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	// Check the response body contains the expected error message
	if !strings.Contains(rr.Body.String(), "Invalid token or missing required scope") {
		t.Errorf("Expected response body to contain %q, got %q", "Invalid token or missing required scope", rr.Body.String())
	}
}

func TestAuthMiddlewareWithMissingToken(t *testing.T) {
	// Create a test configuration with mock values
	testConfig := config.TestConfig(&config.Config{
		Auth: config.AuthConfig{
			KeycloakURL:   "http://mock-keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RequiredScope: "test-scope",
		},
	})

	// Create a mock handler that will be wrapped by the auth middleware
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}

	// Create a mock HTTP client (not used in this test)
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return createValidTokenResponse("test-scope"), nil
		},
	}

	// Create the auth middleware with the test configuration and mock client
	authMiddleware := NewAuthMiddlewareWithClient(testConfig.Auth, mockClient)

	// Create a test request without a token
	req := httptest.NewRequest("GET", "/test", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the middleware with the mock handler
	authMiddleware(mockHandler)(rr, req)

	// Check the response
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	// Check the response body contains the expected error message
	if !strings.Contains(rr.Body.String(), "No token provided") {
		t.Errorf("Expected response body to contain %q, got %q", "No token provided", rr.Body.String())
	}
}

func TestAuthMiddlewareWithInvalidTokenFormat(t *testing.T) {
	// Create a test configuration with mock values
	testConfig := config.TestConfig(&config.Config{
		Auth: config.AuthConfig{
			KeycloakURL:   "http://mock-keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RequiredScope: "test-scope",
		},
	})

	// Create a mock handler that will be wrapped by the auth middleware
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}

	// Create a mock HTTP client (not used in this test)
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return createValidTokenResponse("test-scope"), nil
		},
	}

	// Create the auth middleware with the test configuration and mock client
	authMiddleware := NewAuthMiddlewareWithClient(testConfig.Auth, mockClient)

	// Create a test request with an invalid token format
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the middleware with the mock handler
	authMiddleware(mockHandler)(rr, req)

	// Check the response
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	// Check the response body contains the expected error message
	if !strings.Contains(rr.Body.String(), "Invalid token format") {
		t.Errorf("Expected response body to contain %q, got %q", "Invalid token format", rr.Body.String())
	}
}

func TestAuthMiddlewareWithMissingRequiredScope(t *testing.T) {
	// Create a test configuration with mock values
	testConfig := config.TestConfig(&config.Config{
		Auth: config.AuthConfig{
			KeycloakURL:   "http://mock-keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RequiredScope: "test-scope",
		},
	})

	// Create a mock handler that will be wrapped by the auth middleware
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}

	// Create a mock HTTP client that returns a token missing the required scope
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return createMissingRequiredScopeResponse(), nil
		},
	}

	// Create the auth middleware with the test configuration and mock client
	authMiddleware := NewAuthMiddlewareWithClient(testConfig.Auth, mockClient)

	// Create a test request with a token missing the required scope
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token-missing-scope")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the middleware with the mock handler
	authMiddleware(mockHandler)(rr, req)

	// Check the response
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	// Check the response body contains the expected error message
	if !strings.Contains(rr.Body.String(), "Invalid token or missing required scope") {
		t.Errorf("Expected response body to contain %q, got %q", "Invalid token or missing required scope", rr.Body.String())
	}
}

func TestAuthMiddlewareWithPreflightRequest(t *testing.T) {
	// Create a test configuration with mock values
	testConfig := config.TestConfig(&config.Config{
		Auth: config.AuthConfig{
			KeycloakURL:   "http://mock-keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RequiredScope: "test-scope",
		},
	})

	// Create a mock handler that will be wrapped by the auth middleware
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}

	// Create a mock HTTP client (not used in this test)
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return createValidTokenResponse("test-scope"), nil
		},
	}

	// Create the auth middleware with the test configuration and mock client
	authMiddleware := NewAuthMiddlewareWithClient(testConfig.Auth, mockClient)

	// Create a test preflight request
	req := httptest.NewRequest("OPTIONS", "/test", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the middleware with the mock handler
	authMiddleware(mockHandler)(rr, req)

	// Check the response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Check the CORS headers
	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin header to be %q, got %q", "*", rr.Header().Get("Access-Control-Allow-Origin"))
	}
	if rr.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("Expected Access-Control-Allow-Methods header to be %q, got %q", "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
	}
	if rr.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
		t.Errorf("Expected Access-Control-Allow-Headers header to be %q, got %q", "Content-Type, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestAuthMiddlewareWithMissingClientSecret(t *testing.T) {
	// Create a test configuration with missing client secret
	testConfig := config.TestConfig(&config.Config{
		Auth: config.AuthConfig{
			KeycloakURL:   "http://mock-keycloak:8080",
			ClientID:      "test-client",
			RequiredScope: "test-scope",
			// ClientSecret is intentionally missing
		},
	})

	// Create a mock handler that will be wrapped by the auth middleware
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}

	// Create a mock HTTP client (not used in this test)
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return createValidTokenResponse("test-scope"), nil
		},
	}

	// Create the auth middleware with the test configuration and mock client
	authMiddleware := NewAuthMiddlewareWithClient(testConfig.Auth, mockClient)

	// Create a test request with a valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the middleware with the mock handler
	authMiddleware(mockHandler)(rr, req)

	// Check the response
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	// Check the response body contains the expected error message
	if !strings.Contains(rr.Body.String(), "Token validation failed") {
		t.Errorf("Expected response body to contain %q, got %q", "Token validation failed", rr.Body.String())
	}
}
