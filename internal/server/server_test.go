package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestHealthHandler(t *testing.T) {
	server := New(nil)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status 'ok', got %s", response["status"])
	}

	if response["service"] != "ephd" {
		t.Errorf("expected service 'ephd', got %s", response["service"])
	}
}

func TestEnvironmentsHandler(t *testing.T) {
	server := New(nil)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{"GET environments", "GET", http.StatusOK},
		{"POST environment", "POST", http.StatusNotImplemented},
		{"PUT not allowed", "PUT", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/environments", nil)
			w := httptest.NewRecorder()

			server.environmentsHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	server := New(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := server.applyMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected CORS headers to be set")
	}

	// Check request ID
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected request ID header to be set")
	}
}

func TestStatusHandler(t *testing.T) {
	server := New(nil)

	req := httptest.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()

	server.statusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %v", response["status"])
	}
}

func TestEnvironmentIDHandler(t *testing.T) {
	server := New(nil)

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{"DELETE environment", "/api/v1/environments/test-id", "DELETE", http.StatusNotImplemented},
		{"GET logs", "/api/v1/environments/test-id/logs", "GET", http.StatusNotImplemented},
		{"Invalid path", "/api/v1/environments/test-id", "GET", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			server.environmentIDHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestNotFoundHandler(t *testing.T) {
	server := New(nil)

	req := httptest.NewRequest("GET", "/unknown/path", nil)
	w := httptest.NewRecorder()

	server.notFoundHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["error"] != "Not found" {
		t.Errorf("expected error 'Not found', got %s", response["error"])
	}
}

func TestCORSMiddleware(t *testing.T) {
	server := New(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := server.applyMiddleware(handler)

	// Test OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d for OPTIONS, got %d", http.StatusOK, w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS origin header to be '*'")
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected CORS methods header to be set")
	}

	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("expected CORS headers header to be set")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	// Handler that panics
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("test panic")
	})

	wrapped := recoveryMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic
	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d after panic, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, status: 200}

	// Test default status
	if rw.status != 200 {
		t.Errorf("expected default status 200, got %d", rw.status)
	}

	// Test WriteHeader
	rw.WriteHeader(404)
	if rw.status != 404 {
		t.Errorf("expected status 404, got %d", rw.status)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Port != ":8080" {
		t.Errorf("expected port :8080, got %s", cfg.Port)
	}

	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("expected read timeout 15s, got %v", cfg.ReadTimeout)
	}

	if cfg.WriteTimeout != 15*time.Second {
		t.Errorf("expected write timeout 15s, got %v", cfg.WriteTimeout)
	}

	if cfg.IdleTimeout != 60*time.Second {
		t.Errorf("expected idle timeout 60s, got %v", cfg.IdleTimeout)
	}
}

func TestNewServer(t *testing.T) {
	// Test with nil config
	server := New(nil)
	if server.port != ":8080" {
		t.Errorf("expected default port :8080, got %s", server.port)
	}

	// Test with custom config
	cfg := &Config{Port: ":9090"}
	server = New(cfg)
	if server.port != ":9090" {
		t.Errorf("expected custom port :9090, got %s", server.port)
	}
}

func TestShutdown(t *testing.T) {
	server := New(&Config{Port: ":0"}) // Random port

	// Start server in goroutine
	go func() {
		_ = server.Start()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		t.Errorf("unexpected shutdown error: %v", err)
	}
}

func TestRunWithSignal(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set a custom port via environment variable or config
	os.Setenv("EPH_PORT", ":0")
	defer os.Unsetenv("EPH_PORT")

	// Start server in goroutine
	done := make(chan error, 1)
	go func() {
		done <- Run()
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Send interrupt signal
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}
	if err := p.Signal(syscall.SIGINT); err != nil {
		t.Errorf("failed to send signal: %v", err)
	}

	// Wait for Run to complete
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Run() did not complete within timeout")
	}
}

func TestRunServerError(t *testing.T) {
	// Test Run function when server fails to start
	// This is difficult to test directly without modifying the code
	// to inject errors, so we'll skip this for now
	t.Skip("Testing server start errors requires code refactoring")
}

// Benchmark test
func BenchmarkHealthHandler(b *testing.B) {
	server := New(nil)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		server.healthHandler(w, req)
	}
}

func BenchmarkMiddlewareStack(b *testing.B) {
	server := New(nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := server.applyMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}
}
