package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/handlers"
	"github.com/schneefisch/oauth_keycloak_demo/backend/internal/repository"
)

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Set up database connection
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Error setting up database: %v", err)
	}
	defer db.Close()

	// Create repository
	eventsRepo := repository.NewPostgresEventsRepository(db)

	// Create a new events handler with the repository
	eventsHandler := handlers.NewEventsHandler(eventsRepo)

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

// setupDatabase creates a connection to the PostgreSQL database
func setupDatabase() (*sql.DB, error) {
	// Get database connection details from environment variables
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "postgres"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "bn_keycloak"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "Q6uktXCjQ"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "bitnami_keycloak"
	}

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

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
