package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/repository"
)

// EventsHandler handles HTTP requests related to events
type EventsHandler struct {
	repo repository.EventsRepository
}

// NewEventsHandler creates a new EventsHandler
func NewEventsHandler(repo repository.EventsRepository) *EventsHandler {
	return &EventsHandler{
		repo: repo,
	}
}

// GetEvents returns a list of events from the repository
func (h *EventsHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get events from repository
	events, err := h.repo.GetEvents(context.Background())
	if err != nil {
		http.Error(w, "Error retrieving events", http.StatusInternalServerError)
		return
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Encode events to JSON and write to response
	if err := json.NewEncoder(w).Encode(events); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// GetEventByID returns a specific event by its ID
func (h *EventsHandler) GetEventByID(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the event ID from the URL path using Go 1.22 path variables
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	// Get the event from the repository
	event, err := h.repo.GetEventByID(context.Background(), id)
	if err != nil {
		http.Error(w, "Error retrieving event", http.StatusInternalServerError)
		return
	}

	// If event is nil, it means it wasn't found
	if event == nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Encode event to JSON and write to response
	if err := json.NewEncoder(w).Encode(event); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
