package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/models"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/repository"
)

// TestRoutesSetup tests that the routes are set up correctly
func TestRoutesSetup(t *testing.T) {
	// Create a mock repository
	mockRepo := repository.NewMockEventsRepository()

	// Create a new events handler with the mock repository
	handler := NewEventsHandler(mockRepo)

	// Create a mock auth config
	mockAuthConfig := config.AuthConfig{
		KeycloakURL:   "http://mock-keycloak:8080",
		ClientID:      "test-client",
		ClientSecret:  "test-secret",
		RequiredScope: "test-scope",
	}

	// Create a mock HTTP client
	mockClient := createMockHTTPClient()

	// Create a new test server with the routes set up
	mux := http.NewServeMux()
	// Use the SetupRoutesWithClient function with the mock client
	http.DefaultServeMux = mux
	SetupRoutesWithClient(handler, mockAuthConfig, mockClient)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test cases for different routes
	testCases := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		validateBody   bool
	}{
		{
			name:           "Get all events",
			path:           "/events",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			validateBody:   true,
		},
		{
			name:           "Get event by ID",
			path:           "/events/" + mockRepo.FixedEventID,
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			validateBody:   true,
		},
		{
			name:           "Get non-existent event",
			path:           "/events/non-existent-id",
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound,
			validateBody:   false,
		},
		{
			name:           "Method not allowed for events",
			path:           "/events",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			validateBody:   false,
		},
		{
			name:           "Method not allowed for event by ID",
			path:           "/events/" + mockRepo.FixedEventID,
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			validateBody:   false,
		},
		{
			name:           "Health check",
			path:           "/health",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			validateBody:   false,
		},
		{
			name:           "Redirect from /events/ to /events",
			path:           "/events/",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK, // After redirect
			validateBody:   true,
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new HTTP request with a valid token
			req, err := http.NewRequest(tc.method, server.URL+tc.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Add Authorization header for all requests except health check
			if tc.path != "/health" {
				req.Header.Set("Authorization", "Bearer valid-token")
			}

			// Send the request
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					// Allow redirects
					return nil
				},
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			// Check the status code
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %v, got %v", tc.expectedStatus, resp.StatusCode)
			}

			// If we need to validate the body, do so
			if tc.validateBody {
				// Check the content type
				expectedContentType := "application/json"
				if contentType := resp.Header.Get("Content-Type"); contentType != expectedContentType {
					t.Errorf("Expected content type %v, got %v", expectedContentType, contentType)
				}

				// Decode the response body
				if tc.path == "/events" || tc.path == "/events/" {
					var events models.Events
					if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
						t.Fatalf("Failed to decode response body: %v", err)
					}

					// Check that we got some events
					if len(events) == 0 {
						t.Errorf("Expected events, got empty slice")
					}
				} else if tc.path == "/events/"+mockRepo.FixedEventID {
					var event models.Event
					if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
						t.Fatalf("Failed to decode response body: %v", err)
					}

					// Check that the event has the expected ID
					if event.ID != mockRepo.FixedEventID {
						t.Errorf("Expected event ID %s, got %s", mockRepo.FixedEventID, event.ID)
					}
				}
			}
		})
	}
}
