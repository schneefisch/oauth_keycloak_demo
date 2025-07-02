package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/models"
)

// PostgresEventsRepository implements EventsRepository using PostgreSQL
type PostgresEventsRepository struct {
	db *sql.DB
}

// NewPostgresEventsRepository creates a new PostgresEventsRepository
func NewPostgresEventsRepository(db *sql.DB) *PostgresEventsRepository {
	return &PostgresEventsRepository{
		db: db,
	}
}

// GetEvents retrieves all events from the database
func (r *PostgresEventsRepository) GetEvents(ctx context.Context) (models.Events, error) {
	query := `
		SELECT id, date, title, description, location
		FROM events.events
		ORDER BY date ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events models.Events
	for rows.Next() {
		var event models.Event
		var date time.Time

		if err := rows.Scan(&event.ID, &date, &event.Title, &event.Description, &event.Location); err != nil {
			return nil, err
		}

		event.Date = date
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
