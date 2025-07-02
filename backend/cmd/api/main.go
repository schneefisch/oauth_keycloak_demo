package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/handlers"
)

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create a new events handler
	eventsHandler := handlers.NewEventsHandler()

	// Register the handler for the /events endpoint
	http.HandleFunc("/events", eventsHandler.GetEvents)

	// Add a simple health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start the server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}