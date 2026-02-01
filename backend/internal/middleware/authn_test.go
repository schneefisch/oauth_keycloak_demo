package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/oauth"
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
	response := oauth.TokenIntrospectionResponse{
		Active:    true,
		Scope:     requiredScope,
		ClientID:  "test-client",
		Username:  "test-user",
		TokenType: "Bearer",
		Exp:       1735689600,
		Iat:       1619712000,
		Nbf:       1619712000,
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
	response := oauth.TokenIntrospectionResponse{
		Active: false,
	}

	responseBody, _ := json.Marshal(response)
	return createMockResponse(http.StatusOK, string(responseBody))
}

// createErrorResponse creates a mock response for an error
func createErrorResponse() *http.Response {
	return createMockResponse(http.StatusInternalServerError, "Internal Server Error")
}

func TestAuthMiddleware_NoAuthorizationHeader(t *testing.T) {
	testConfig := config.TestConfig(&config.Config{
		Auth: config.AuthConfig{
			KeycloakURL:   "http://mock-keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RequiredScope: "test-scope",
			RealmName:     "test-realm",
		},
	})

	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return createValidTokenResponse("test-scope"), nil
		},
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := NewIntrospectionAuthMiddlewareWithClient(testConfig.Auth, mockClient)
	handler := authMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if handlerCalled {
		t.Error("Handler should not have been called")
	}
}

func TestAuthMiddleware_InvalidAuthorizationHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"empty header", ""},
		{"no bearer prefix", "token123"},
		{"bearer only", "Bearer"},
		{"bearer with empty token", "Bearer "},
		{"wrong prefix", "Basic token123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testConfig := config.TestConfig(&config.Config{
				Auth: config.AuthConfig{
					KeycloakURL:   "http://mock-keycloak:8080",
					ClientID:      "test-client",
					ClientSecret:  "test-secret",
					RequiredScope: "test-scope",
					RealmName:     "test-realm",
				},
			})

			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return createValidTokenResponse("test-scope"), nil
				},
			}

			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			authMiddleware := NewIntrospectionAuthMiddlewareWithClient(testConfig.Auth, mockClient)
			handler := authMiddleware(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
			}
			if handlerCalled {
				t.Error("Handler should not have been called")
			}
		})
	}
}

func TestAuthMiddleware_ValidToken_StoresClaimsInContext(t *testing.T) {
	testConfig := config.TestConfig(&config.Config{
		Auth: config.AuthConfig{
			KeycloakURL:   "http://mock-keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RequiredScope: "test-scope events:read",
			RealmName:     "test-realm",
		},
	})

	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify the introspection request
			if strings.Contains(req.URL.Path, "token/introspect") {
				return createValidTokenResponse("test-scope events:read"), nil
			}
			return createErrorResponse(), nil
		},
	}

	var capturedClaims *oauth.AuthClaims
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = oauth.GetAuthClaims(r)
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := NewIntrospectionAuthMiddlewareWithClient(testConfig.Auth, mockClient)
	handler := authMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if capturedClaims == nil {
		t.Fatal("Expected claims to be stored in context")
	}
	if capturedClaims.Subject != "test-subject" {
		t.Errorf("Expected Subject 'test-subject', got '%s'", capturedClaims.Subject)
	}
	if capturedClaims.Username != "test-user" {
		t.Errorf("Expected Username 'test-user', got '%s'", capturedClaims.Username)
	}
	if !capturedClaims.HasScope("test-scope") {
		t.Error("Expected claims to have 'test-scope'")
	}
	if !capturedClaims.HasScope("events:read") {
		t.Error("Expected claims to have 'events:read'")
	}
}

func TestAuthMiddleware_InactiveToken(t *testing.T) {
	testConfig := config.TestConfig(&config.Config{
		Auth: config.AuthConfig{
			KeycloakURL:   "http://mock-keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RequiredScope: "test-scope",
			RealmName:     "test-realm",
		},
	})

	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return createInvalidTokenResponse(), nil
		},
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := NewIntrospectionAuthMiddlewareWithClient(testConfig.Auth, mockClient)
	handler := authMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer inactive-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if handlerCalled {
		t.Error("Handler should not have been called")
	}
}

func TestAuthMiddleware_IntrospectionError(t *testing.T) {
	testConfig := config.TestConfig(&config.Config{
		Auth: config.AuthConfig{
			KeycloakURL:   "http://mock-keycloak:8080",
			ClientID:      "test-client",
			ClientSecret:  "test-secret",
			RequiredScope: "test-scope",
			RealmName:     "test-realm",
		},
	})

	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return createErrorResponse(), nil
		},
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := NewIntrospectionAuthMiddlewareWithClient(testConfig.Auth, mockClient)
	handler := authMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if handlerCalled {
		t.Error("Handler should not have been called")
	}
}
