package models

import (
	"time"
)

// Event represents an event in the system
type Event struct {
	ID          string    `json:"id"`
	Date        time.Time `json:"date"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
}

// Events is a slice of Event
type Events []Event
