package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

// Appointment represents a training appointment
type Appointment struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
}

// AppointmentStore is a simple in-memory store for appointments
type AppointmentStore struct {
	sync.RWMutex
	appointments map[string]Appointment
}

// Global store
var store = AppointmentStore{
	appointments: make(map[string]Appointment),
}

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Define routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/appointments", appointmentsHandler)
	http.HandleFunc("/api/appointments/", appointmentHandler)

	// Start server
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// appointmentsHandler handles requests to /api/appointments
func appointmentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Return all appointments
		store.RLock()
		appointments := make([]Appointment, 0, len(store.appointments))
		for _, appointment := range store.appointments {
			appointments = append(appointments, appointment)
		}
		store.RUnlock()

		if err := json.NewEncoder(w).Encode(appointments); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		// Create a new appointment
		var appointment Appointment
		if err := json.NewDecoder(r.Body).Decode(&appointment); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		store.Lock()
		store.appointments[appointment.ID] = appointment
		store.Unlock()

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(appointment); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// appointmentHandler handles requests to /api/appointments/{id}
func appointmentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract appointment ID from URL
	id := r.URL.Path[len("/api/appointments/"):]
	if id == "" {
		http.Error(w, "Invalid appointment ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Get appointment by ID
		store.RLock()
		appointment, exists := store.appointments[id]
		store.RUnlock()

		if !exists {
			http.Error(w, "Appointment not found", http.StatusNotFound)
			return
		}

		if err := json.NewEncoder(w).Encode(appointment); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case http.MethodPut:
		// Update appointment
		var appointment Appointment
		if err := json.NewDecoder(r.Body).Decode(&appointment); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		appointment.ID = id

		store.Lock()
		store.appointments[id] = appointment
		store.Unlock()

		if err := json.NewEncoder(w).Encode(appointment); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case http.MethodDelete:
		// Delete appointment
		store.Lock()
		delete(store.appointments, id)
		store.Unlock()

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}