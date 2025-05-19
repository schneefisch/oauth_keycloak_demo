package models

import (
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
	Appointments map[string]Appointment
}

// NewAppointmentStore creates a new appointment store
func NewAppointmentStore() *AppointmentStore {
	return &AppointmentStore{
		Appointments: make(map[string]Appointment),
	}
}

// GetAll returns all appointments
func (s *AppointmentStore) GetAll() []Appointment {
	s.RLock()
	defer s.RUnlock()

	appointments := make([]Appointment, 0, len(s.Appointments))
	for _, appointment := range s.Appointments {
		appointments = append(appointments, appointment)
	}
	return appointments
}

// Get returns an appointment by ID
func (s *AppointmentStore) Get(id string) (Appointment, bool) {
	s.RLock()
	defer s.RUnlock()

	appointment, exists := s.Appointments[id]
	return appointment, exists
}

// Create adds a new appointment
func (s *AppointmentStore) Create(appointment Appointment) {
	s.Lock()
	defer s.Unlock()

	s.Appointments[appointment.ID] = appointment
}

// Update updates an existing appointment
func (s *AppointmentStore) Update(id string, appointment Appointment) {
	s.Lock()
	defer s.Unlock()

	s.Appointments[id] = appointment
}

// Delete removes an appointment
func (s *AppointmentStore) Delete(id string) {
	s.Lock()
	defer s.Unlock()

	delete(s.Appointments, id)
}
