package handlers

import (
	"net/http"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/middleware"
)

// SetupRoutes configures all the HTTP routes for the application
func SetupRoutes(eventsHandler *EventsHandler, authConfig config.AuthConfig) {
	SetupRoutesWithClient(eventsHandler, authConfig, &http.Client{})
}

// SetupRoutesWithClient configures all the HTTP routes for the application with a custom HTTP client
func SetupRoutesWithClient(eventsHandler *EventsHandler, authConfig config.AuthConfig, client middleware.HTTPClient) {
	// Create AuthN middleware (validates token, stores claims in context)
	authN := middleware.NewAuthMiddlewareWithClient(authConfig, client)

	// Create AuthZ middleware (checks required scopes)
	// Using the required scope from config for backward compatibility
	authZ := middleware.NewAuthzMiddleware(middleware.AuthzConfig{
		RequiredScopes: []string{authConfig.RequiredScope},
		RequireAll:     true,
	})

	// Helper function to chain AuthN and AuthZ middlewares
	withAuth := func(handler http.HandlerFunc) http.HandlerFunc {
		return authN(authZ(handler))
	}

	// Register the handler for the /events/{id} endpoint using Go 1.22 path variables
	// This must be registered first as it's more specific than "/events/"
	http.HandleFunc("/events/{id}", withAuth(eventsHandler.GetEventByID))

	// Register the handler for the /events endpoint
	http.HandleFunc("/events", withAuth(eventsHandler.GetEvents))

	// Handle the specific case of "/events/" to redirect to "/events"
	http.HandleFunc("/events/", withAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/events/" {
			http.Redirect(w, r, "/events", http.StatusMovedPermanently)
			return
		}
	}))

	// Add a simple health check endpoint (no auth required)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}
