package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/config"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/handlers"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/repository"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Set up database connection
	db, err := setupDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Error setting up database: %v", err)
	}
	defer db.Close()

	// Create repository
	eventsRepo := repository.NewPostgresEventsRepository(db)

	// Create a new events handler with the repository
	eventsHandler := handlers.NewEventsHandler(eventsRepo)

	// Setup all routes with auth configuration
	handlers.SetupRoutes(eventsHandler, cfg.Auth)

	// Start the server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

// setupDatabase creates a connection to the PostgreSQL database
func setupDatabase(dbConfig config.DatabaseConfig) (*sql.DB, error) {
	// Create connection string using the configuration
	connStr := dbConfig.ConnectionString()

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Successfully connected to database")
	return db, nil
}
