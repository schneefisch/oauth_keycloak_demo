package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"../models"
)

// AppointmentHandler contains handlers for appointment endpoints
type AppointmentHandler struct {
	Store *models.AppointmentStore
}

// NewAppointmentHandler creates a new appointment handler
func NewAppointmentHandler(store *models.AppointmentStore) *AppointmentHandler {
	return &AppointmentHandler{
		Store: store,
	}
}

// ListCreate handles GET and POST requests to /api/appointments
func (h *AppointmentHandler) ListCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Return all appointments
		appointments := h.Store.GetAll()

		if err := json.NewEncoder(w).Encode(appointments); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		// Create a new appointment
		var appointment models.Appointment
		if err := json.NewDecoder(r.Body).Decode(&appointment); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		h.Store.Create(appointment)

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(appointment); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetUpdateDelete handles GET, PUT and DELETE requests to /api/appointments/{id}
func (h *AppointmentHandler) GetUpdateDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract appointment ID from URL
	path := r.URL.Path
	id := strings.TrimPrefix(path, "/api/appointments/")

	if id == "" {
		http.Error(w, "Invalid appointment ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Get appointment by ID
		appointment, exists := h.Store.Get(id)

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
		var appointment models.Appointment
		if err := json.NewDecoder(r.Body).Decode(&appointment); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		appointment.ID = id
		h.Store.Update(id, appointment)

		if err := json.NewEncoder(w).Encode(appointment); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case http.MethodDelete:
		// Delete appointment
		h.Store.Delete(id)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
