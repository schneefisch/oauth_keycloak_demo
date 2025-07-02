package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/models"
)

// EventsHandler handles HTTP requests related to events
type EventsHandler struct{}

// NewEventsHandler creates a new EventsHandler
func NewEventsHandler() *EventsHandler {
	return &EventsHandler{}
}

// GetEvents returns a list of events
// For simplicity, this returns mock data
func (h *EventsHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create some mock events
	events := models.Events{
		{
			ID:          uuid.New().String(),
			Date:        time.Now().AddDate(0, 0, 7),
			Title:       "Community Soccer Match",
			Description: "Weekly soccer match for all community members",
			Location:    "Community Field",
		},
		{
			ID:          uuid.New().String(),
			Date:        time.Now().AddDate(0, 0, 14),
			Title:       "Basketball Tournament",
			Description: "Annual basketball tournament with teams from neighboring communities",
			Location:    "Sports Center",
		},
		{
			ID:          uuid.New().String(),
			Date:        time.Now().AddDate(0, 0, 21),
			Title:       "Swimming Competition",
			Description: "Swimming competition for all age groups",
			Location:    "Community Pool",
		},
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Encode events to JSON and write to response
	if err := json.NewEncoder(w).Encode(events); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
