package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware_SetsHeaders(t *testing.T) {
	config := DefaultCORSConfig()
	corsMiddleware := NewCORSMiddleware(config)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := corsMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handlerCalled {
		t.Error("Handler should have been called")
	}

	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}

	for header, expected := range expectedHeaders {
		if got := rr.Header().Get(header); got != expected {
			t.Errorf("Expected header %s to be '%s', got '%s'", header, expected, got)
		}
	}
}

func TestCORSMiddleware_HandlesPreflight(t *testing.T) {
	config := DefaultCORSConfig()
	corsMiddleware := NewCORSMiddleware(config)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	handler := corsMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if handlerCalled {
		t.Error("Handler should not be called for preflight requests")
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS header to be set")
	}
}

func TestCORSMiddleware_CustomConfig(t *testing.T) {
	config := CORSConfig{
		AllowedOrigins: []string{"http://example.com"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"X-Custom-Header"},
	}
	corsMiddleware := NewCORSMiddleware(config)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := corsMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "http://example.com",
		"Access-Control-Allow-Methods": "GET, POST",
		"Access-Control-Allow-Headers": "X-Custom-Header",
	}

	for header, expected := range expectedHeaders {
		if got := rr.Header().Get(header); got != expected {
			t.Errorf("Expected header %s to be '%s', got '%s'", header, expected, got)
		}
	}
}

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	if len(config.AllowedOrigins) != 1 || config.AllowedOrigins[0] != "*" {
		t.Errorf("Expected AllowedOrigins to be ['*'], got %v", config.AllowedOrigins)
	}
	if len(config.AllowedMethods) != 5 {
		t.Errorf("Expected 5 allowed methods, got %d", len(config.AllowedMethods))
	}
	if len(config.AllowedHeaders) != 2 {
		t.Errorf("Expected 2 allowed headers, got %d", len(config.AllowedHeaders))
	}
}
