package handlers

import (
	"log"
	"net/http"
	"strings"
)

// AuthMiddleware is a middleware that logs the received token
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// The Authorization header typically has the format "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token := parts[1]
				// Log the token
				log.Printf("Received token: %s", token)
			} else {
				log.Printf("Invalid Authorization header format: %s", authHeader)
			}
		} else {
			log.Printf("No Authorization header found in request to %s", r.URL.Path)
		}

		// Set CORS headers to allow requests from the frontend
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next(w, r)
	}
}
