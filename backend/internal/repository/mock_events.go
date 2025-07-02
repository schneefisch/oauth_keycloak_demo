package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/models"
)

// MockEventsRepository implements EventsRepository for testing
type MockEventsRepository struct {
	// Store a fixed event ID for testing GetEventByID
	FixedEventID string
}

// NewMockEventsRepository creates a new MockEventsRepository
func NewMockEventsRepository() *MockEventsRepository {
	return &MockEventsRepository{
		FixedEventID: "event-123", // Fixed ID for testing
	}
}

// GetEvents returns mock events for testing
func (r *MockEventsRepository) GetEvents(ctx context.Context) (models.Events, error) {
	// Create some mock events
	events := models.Events{
		{
			ID:          r.FixedEventID, // Use the fixed ID for the first event
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

	return events, nil
}

// GetEventByID returns a mock event for testing
func (r *MockEventsRepository) GetEventByID(ctx context.Context, id string) (*models.Event, error) {
	// If the ID matches our fixed ID, return a mock event
	if id == r.FixedEventID {
		return &models.Event{
			ID:          r.FixedEventID,
			Date:        time.Now().AddDate(0, 0, 7),
			Title:       "Community Soccer Match",
			Description: "Weekly soccer match for all community members",
			Location:    "Community Field",
		}, nil
	}

	// If the ID doesn't match, return nil to simulate not found
	return nil, nil
}
