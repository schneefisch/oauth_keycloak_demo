package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/models"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/repository"
)

func TestGetEvents(t *testing.T) {
	// Create a mock repository
	mockRepo := repository.NewMockEventsRepository()

	// Create a new events handler with the mock repository
	handler := NewEventsHandler(mockRepo)

	// Create a new test server with the routes set up
	mux := http.NewServeMux()
	http.DefaultServeMux = mux
	SetupRoutes(handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Create a new HTTP request
	resp, err := http.Get(server.URL + "/events")
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

	// Create a new test server with the routes set up
	mux := http.NewServeMux()
	http.DefaultServeMux = mux
	SetupRoutes(handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Create a new HTTP request with POST method
	req, err := http.NewRequest(http.MethodPost, server.URL+"/events", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

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

	// Create a new test server with the routes set up
	mux := http.NewServeMux()
	http.DefaultServeMux = mux
	SetupRoutes(handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Send a GET request to the server
	resp, err := http.Get(server.URL + "/events/" + mockRepo.FixedEventID)
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

	// Create a new test server with the routes set up
	mux := http.NewServeMux()
	http.DefaultServeMux = mux
	SetupRoutes(handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Send a GET request to the server with a non-existent event ID
	resp, err := http.Get(server.URL + "/events/non-existent-id")
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

	// Create a new test server with the routes set up
	mux := http.NewServeMux()
	http.DefaultServeMux = mux
	SetupRoutes(handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Create a new HTTP request with POST method
	req, err := http.NewRequest(http.MethodPost, server.URL+"/events/"+mockRepo.FixedEventID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

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
