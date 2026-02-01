package handlers

import (
	"net/http"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/middleware"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/oauth"
)

// SetupRoutes configures all the HTTP routes for the application
func SetupRoutes(eventsHandler *EventsHandler, authConfig config.AuthConfig) {
	SetupRoutesWithClient(eventsHandler, authConfig, &http.Client{})
}

// SetupRoutesWithClient configures all the HTTP routes for the application with a custom HTTP client
func SetupRoutesWithClient(eventsHandler *EventsHandler, authConfig config.AuthConfig, client oauth.HTTPClient) {
	// Create CORS middleware
	cors := middleware.NewCORSMiddleware(middleware.DefaultCORSConfig())

	// Create AuthN middleware (validates token, stores claims in context)
	authN := middleware.NewAuthMiddlewareWithClient(authConfig, client)

	// Create AuthZ middleware (checks required scopes)
	authZ := middleware.NewAuthzMiddleware(middleware.AuthzConfig{
		RequiredScopes: []string{authConfig.RequiredScope},
		RequireAll:     true,
	})

	// Helper function to chain CORS -> AuthN -> AuthZ middlewares
	protected := func(h http.Handler) http.Handler {
		return cors(authN(authZ(h)))
	}

	// Register the handler for the /events/{id} endpoint using Go 1.22 path variables
	// This must be registered first as it's more specific than "/events/"
	http.Handle("/events/{id}", protected(http.HandlerFunc(eventsHandler.GetEventByID)))

	// Register the handler for the /events endpoint
	http.Handle("/events", protected(http.HandlerFunc(eventsHandler.GetEvents)))

	// Handle the specific case of "/events/" to redirect to "/events"
	http.Handle("/events/", protected(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/events/" {
			http.Redirect(w, r, "/events", http.StatusMovedPermanently)
			return
		}
	})))

	// Add a simple health check endpoint (CORS only, no auth required)
	http.Handle("/health", cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})))
}
