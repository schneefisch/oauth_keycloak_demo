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
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/middleware"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/models"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/oauth"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/repository"
)

// MockHTTPClient is a mock implementation of the HTTPClient interface
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do implements the HTTPClient interface
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// createMockAuthConfig creates a mock auth config for testing
func createMockAuthConfig() config.AuthConfig {
	return config.AuthConfig{
		KeycloakURL:   "http://mock-keycloak:8080",
		ClientID:      "test-client",
		ClientSecret:  "test-secret",
		RequiredScope: "test-scope",
	}
}

// createMockHTTPClient creates a mock HTTP client that returns a valid token response
func createMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Check if this is a token introspection request
			if strings.Contains(req.URL.Path, "token/introspect") {
				// Return a valid token response
				response := oauth.TokenIntrospectionResponse{
					Active:    true,
					Scope:     "test-scope",
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
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(string(responseBody))),
					Header:     make(http.Header),
				}, nil
			}

			// For any other request, return a generic success response
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("OK")),
				Header:     make(http.Header),
			}, nil
		},
	}
}

func TestGetEvents(t *testing.T) {
	// Create a mock repository
	mockRepo := repository.NewMockEventsRepository()

	// Create a new events handler with the mock repository
	handler := NewEventsHandler(mockRepo)

	// Create a mock auth config
	mockAuthConfig := createMockAuthConfig()

	// Create a mock HTTP client
	mockClient := createMockHTTPClient()

	// Create a new test server with the routes set up
	mux := http.NewServeMux()

	// Create the auth middleware with the mock client
	authMiddleware := middleware.NewAuthMiddlewareWithClient(mockAuthConfig, mockClient)

	// Register the routes manually using http.Handler pattern
	mux.Handle("/events/{id}", authMiddleware(http.HandlerFunc(handler.GetEventByID)))
	mux.Handle("/events", authMiddleware(http.HandlerFunc(handler.GetEvents)))
	mux.Handle("/events/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/events/" {
			http.Redirect(w, r, "/events", http.StatusMovedPermanently)
			return
		}
	})))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// Create a new HTTP request with a valid token
	req, err := http.NewRequest("GET", server.URL+"/events", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer valid-token")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the status code
	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the content type
	expectedContentType := "application/json"
	if contentType := resp.Header.Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Handler returned wrong content type: got %v want %v", contentType, expectedContentType)
	}

	// Decode the response body
	var events models.Events
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Check that we got some events
	if len(events) == 0 {
		t.Errorf("Expected events, got empty slice")
	}

	// Check that each event has the required fields
	for i, event := range events {
		if event.ID == "" {
			t.Errorf("Event %d missing ID", i)
		}
		if event.Title == "" {
			t.Errorf("Event %d missing Title", i)
		}
		if event.Description == "" {
			t.Errorf("Event %d missing Description", i)
		}
		if event.Location == "" {
			t.Errorf("Event %d missing Location", i)
		}
		// Date is automatically set, so we don't need to check it
	}
}

// Test that non-GET methods are rejected
func TestGetEventsMethodNotAllowed(t *testing.T) {
	// Create a mock repository
	mockRepo := repository.NewMockEventsRepository()

	// Create a new events handler with the mock repository
	handler := NewEventsHandler(mockRepo)

	// Create a mock auth config
	mockAuthConfig := createMockAuthConfig()

	// Create a mock HTTP client
	mockClient := createMockHTTPClient()

	// Create a new test server with the routes set up
	mux := http.NewServeMux()

	// Create the auth middleware with the mock client
	authMiddleware := middleware.NewAuthMiddlewareWithClient(mockAuthConfig, mockClient)

	// Register the routes manually using http.Handler pattern
	mux.Handle("/events/{id}", authMiddleware(http.HandlerFunc(handler.GetEventByID)))
	mux.Handle("/events", authMiddleware(http.HandlerFunc(handler.GetEvents)))
	mux.Handle("/events/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/events/" {
			http.Redirect(w, r, "/events", http.StatusMovedPermanently)
			return
		}
	})))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// Create a new HTTP request with POST method and a valid token
	req, err := http.NewRequest(http.MethodPost, server.URL+"/events", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer valid-token")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the status code
	if status := resp.StatusCode; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

// Test getting a specific event by ID
func TestGetEventByID(t *testing.T) {
	// Create a mock repository
	mockRepo := repository.NewMockEventsRepository()

	// Create a new events handler with the mock repository
	handler := NewEventsHandler(mockRepo)

	// Create a mock auth config
	mockAuthConfig := createMockAuthConfig()

	// Create a mock HTTP client
	mockClient := createMockHTTPClient()

	// Create a new test server with the routes set up
	mux := http.NewServeMux()

	// Create the auth middleware with the mock client
	authMiddleware := middleware.NewAuthMiddlewareWithClient(mockAuthConfig, mockClient)

	// Register the routes manually using http.Handler pattern
	mux.Handle("/events/{id}", authMiddleware(http.HandlerFunc(handler.GetEventByID)))
	mux.Handle("/events", authMiddleware(http.HandlerFunc(handler.GetEvents)))
	mux.Handle("/events/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/events/" {
			http.Redirect(w, r, "/events", http.StatusMovedPermanently)
			return
		}
	})))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// Create a new HTTP request with a valid token
	req, err := http.NewRequest("GET", server.URL+"/events/"+mockRepo.FixedEventID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer valid-token")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the status code
	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the content type
	expectedContentType := "application/json"
	if contentType := resp.Header.Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Handler returned wrong content type: got %v want %v", contentType, expectedContentType)
	}

	// Decode the response body
	var event models.Event
	if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Check that the event has the expected ID
	if event.ID != mockRepo.FixedEventID {
		t.Errorf("Expected event ID %s, got %s", mockRepo.FixedEventID, event.ID)
	}

	// Check that the event has the required fields
	if event.Title == "" {
		t.Errorf("Event missing Title")
	}
	if event.Description == "" {
		t.Errorf("Event missing Description")
	}
	if event.Location == "" {
		t.Errorf("Event missing Location")
	}
}

// Test getting a non-existent event
func TestGetEventByIDNotFound(t *testing.T) {
	// Create a mock repository
	mockRepo := repository.NewMockEventsRepository()

	// Create a new events handler with the mock repository
	handler := NewEventsHandler(mockRepo)

	// Create a mock auth config
	mockAuthConfig := createMockAuthConfig()

	// Create a mock HTTP client
	mockClient := createMockHTTPClient()

	// Create a new test server with the routes set up
	mux := http.NewServeMux()

	// Create the auth middleware with the mock client
	authMiddleware := middleware.NewAuthMiddlewareWithClient(mockAuthConfig, mockClient)

	// Register the routes manually using http.Handler pattern
	mux.Handle("/events/{id}", authMiddleware(http.HandlerFunc(handler.GetEventByID)))
	mux.Handle("/events", authMiddleware(http.HandlerFunc(handler.GetEvents)))
	mux.Handle("/events/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/events/" {
			http.Redirect(w, r, "/events", http.StatusMovedPermanently)
			return
		}
	})))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// Create a new HTTP request with a valid token
	req, err := http.NewRequest("GET", server.URL+"/events/non-existent-id", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer valid-token")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the status code
	if status := resp.StatusCode; status != http.StatusNotFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

// Test that non-GET methods are rejected for GetEventByID
func TestGetEventByIDMethodNotAllowed(t *testing.T) {
	// Create a mock repository
	mockRepo := repository.NewMockEventsRepository()

	// Create a new events handler with the mock repository
	handler := NewEventsHandler(mockRepo)

	// Create a mock auth config
	mockAuthConfig := createMockAuthConfig()

	// Create a mock HTTP client
	mockClient := createMockHTTPClient()

	// Create a new test server with the routes set up
	mux := http.NewServeMux()

	// Create the auth middleware with the mock client
	authMiddleware := middleware.NewAuthMiddlewareWithClient(mockAuthConfig, mockClient)

	// Register the routes manually using http.Handler pattern
	mux.Handle("/events/{id}", authMiddleware(http.HandlerFunc(handler.GetEventByID)))
	mux.Handle("/events", authMiddleware(http.HandlerFunc(handler.GetEvents)))
	mux.Handle("/events/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/events/" {
			http.Redirect(w, r, "/events", http.StatusMovedPermanently)
			return
		}
	})))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// Create a new HTTP request with POST method and a valid token
	req, err := http.NewRequest(http.MethodPost, server.URL+"/events/"+mockRepo.FixedEventID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer valid-token")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the status code
	if status := resp.StatusCode; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}
