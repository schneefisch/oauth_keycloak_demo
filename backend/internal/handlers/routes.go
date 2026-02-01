package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/middleware"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/oauth"
)

// SetupRoutes configures all the HTTP routes for the application
func SetupRoutes(eventsHandler *EventsHandler, authConfig config.AuthConfig) {
	SetupRoutesWithContext(context.Background(), eventsHandler, authConfig)
}

// SetupRoutesWithContext configures all the HTTP routes with a context for validator lifecycle
// An optional HTTPClient can be provided for testing purposes
func SetupRoutesWithContext(ctx context.Context, eventsHandler *EventsHandler, authConfig config.AuthConfig, client ...oauth.HTTPClient) {
	// Create CORS middleware
	cors := middleware.NewCORSMiddleware(middleware.DefaultCORSConfig())

	// Use provided client or default to http.Client
	var httpClient oauth.HTTPClient = &http.Client{}
	if len(client) > 0 && client[0] != nil {
		httpClient = client[0]
	}

	// Create validator based on config
	method := oauth.ValidationMethod(authConfig.ValidationMethod)
	validator, err := oauth.NewTokenValidator(method, oauth.ValidatorConfig{
		AuthConfig: authConfig,
		HTTPClient: httpClient,
		Context:    ctx,
	})
	if err != nil {
		log.Fatalf("Failed to create token validator: %v", err)
	}

	// Create AuthN middleware using the validator
	authN := middleware.NewAuthMiddlewareWithValidator(validator)

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
