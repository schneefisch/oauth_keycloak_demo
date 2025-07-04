package handlers

import (
	"net/http"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
)

// SetupRoutes configures all the HTTP routes for the application
func SetupRoutes(eventsHandler *EventsHandler, authConfig config.AuthConfig) {
	// Create an auth middleware with the configuration
	authMiddleware := NewAuthMiddleware(authConfig)

	// Register the handler for the /events/{id} endpoint using Go 1.22 path variables
	// This must be registered first as it's more specific than "/events/"
	http.HandleFunc("/events/{id}", authMiddleware(eventsHandler.GetEventByID))

	// Register the handler for the /events endpoint
	http.HandleFunc("/events", authMiddleware(eventsHandler.GetEvents))

	// Handle the specific case of "/events/" to redirect to "/events"
	http.HandleFunc("/events/", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
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

// SetupRoutesWithClient configures all the HTTP routes for the application with a custom HTTP client
func SetupRoutesWithClient(eventsHandler *EventsHandler, authConfig config.AuthConfig, client HTTPClient) {
	// Create an auth middleware with the configuration and custom client
	authMiddleware := NewAuthMiddlewareWithClient(authConfig, client)

	// Register the handler for the /events/{id} endpoint using Go 1.22 path variables
	// This must be registered first as it's more specific than "/events/"
	http.HandleFunc("/events/{id}", authMiddleware(eventsHandler.GetEventByID))

	// Register the handler for the /events endpoint
	http.HandleFunc("/events", authMiddleware(eventsHandler.GetEvents))

	// Handle the specific case of "/events/" to redirect to "/events"
	http.HandleFunc("/events/", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
