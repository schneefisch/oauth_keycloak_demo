package main

import (
	"log"
	"net/http"
	"os"

	"../internal/handlers"
	"../internal/models"
)

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize store
	store := models.NewAppointmentStore()

	// Initialize handlers
	appointmentHandler := handlers.NewAppointmentHandler(store)

	// Define routes
	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/api/appointments", appointmentHandler.ListCreate)
	http.HandleFunc("/api/appointments/", appointmentHandler.GetUpdateDelete)

	// Start server
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
