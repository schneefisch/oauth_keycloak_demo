package handlers

import (
	"net/http"
)

// SetupRoutes configures all the HTTP routes for the application
func SetupRoutes(eventsHandler *EventsHandler) {
	// Register the handler for the /events/{id} endpoint using Go 1.22 path variables
	// This must be registered first as it's more specific than "/events/"
	http.HandleFunc("/events/{id}", AuthMiddleware(eventsHandler.GetEventByID))

	// Register the handler for the /events endpoint
	http.HandleFunc("/events", AuthMiddleware(eventsHandler.GetEvents))

	// Handle the specific case of "/events/" to redirect to "/events"
	http.HandleFunc("/events/", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/events/" {
			http.Redirect(w, r, "/events", http.StatusMovedPermanently)
			return
		}
		// For any other path, the ServeMux will route to the more specific handler
	}))

	// Add a simple health check endpoint (no auth required)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}
