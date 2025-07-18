package repository

import (
	"context"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/models"
)

// EventsRepository defines the interface for event data operations
type EventsRepository interface {
	// GetEvents retrieves all events
	GetEvents(ctx context.Context) (models.Events, error)

	// GetEventByID retrieves a specific event by its ID
	GetEventByID(ctx context.Context, id string) (*models.Event, error)
}
