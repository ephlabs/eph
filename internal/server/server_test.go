package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/ephlabs/eph/internal/log"
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

	if response["version"] == "" {
		t.Error("expected version to be non-empty")
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

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected CORS headers to be set")
	}

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

	if response["version"] == "" {
		t.Error("expected version to be non-empty")
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

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d for OPTIONS, got %d", http.StatusOK, w.Code)
	}

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
	server := New(nil)
	if server.config.Port != ":8080" {
		t.Errorf("expected default port :8080, got %s", server.config.Port)
	}

	cfg := &Config{
		Port:         ":9090",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	server = New(cfg)
	if server.config.Port != ":9090" {
		t.Errorf("expected custom port :9090, got %s", server.config.Port)
	}
	if server.config.ReadTimeout != 10*time.Second {
		t.Errorf("expected custom read timeout 10s, got %v", server.config.ReadTimeout)
	}
	if server.config.WriteTimeout != 20*time.Second {
		t.Errorf("expected custom write timeout 20s, got %v", server.config.WriteTimeout)
	}
	if server.config.IdleTimeout != 30*time.Second {
		t.Errorf("expected custom idle timeout 30s, got %v", server.config.IdleTimeout)
	}
}

func TestShutdown(t *testing.T) {
	server := New(&Config{Port: ":0"}) // Random port

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Start()
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

	// Wait for server to actually stop
	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("unexpected server error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("server did not stop within timeout")
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

// Tests for structured logging implementation

func TestServerStartLogging(t *testing.T) {
	// Save original logger
	originalLogger := log.Default()
	defer log.SetDefault(originalLogger)

	// Capture logs
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{}))
	log.SetDefault(logger)

	server := New(&Config{Port: ":0"}) // Use port 0 for testing

	// Start server in goroutine since it blocks
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Start()
	}()

	// Give server time to start and log
	time.Sleep(100 * time.Millisecond)

	// Shutdown immediately
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)

	// Wait for server to stop
	select {
	case <-serverErr:
	case <-time.After(2 * time.Second):
		t.Error("Server did not stop in time")
	}

	// Check logs
	logOutput := strings.TrimSpace(buf.String())
	if logOutput == "" {
		t.Fatal("Expected server start to be logged")
	}

	// There might be multiple log lines from middleware, get the first one
	lines := strings.Split(logOutput, "\n")
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	// Verify structured log fields
	if msg, ok := logEntry["msg"]; !ok || !strings.Contains(msg.(string), "Starting") {
		t.Errorf("Expected startup message in log, got: %v", logEntry)
	}

	if port, ok := logEntry["port"]; !ok || port != ":0" {
		t.Errorf("Expected port=':0' in log, got: %v", port)
	}
}

func TestJSONResponseLogging(t *testing.T) {
	// Capture logs
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{}))
	log.SetDefault(logger)

	server := New(nil)

	// Test successful JSON encoding
	t.Run("successful encoding", func(t *testing.T) {
		buf.Reset()
		w := httptest.NewRecorder()
		data := map[string]string{"message": "success"}

		server.jsonResponse(w, http.StatusOK, data)

		// Should not log anything on success
		if buf.Len() > 0 {
			t.Error("Expected no logs on successful JSON encoding")
		}
	})

	// Test JSON encoding error
	t.Run("encoding error", func(t *testing.T) {
		buf.Reset()
		w := httptest.NewRecorder()

		// Create data that will fail JSON encoding
		badData := make(chan int) // channels can't be JSON encoded

		server.jsonResponse(w, http.StatusOK, badData)

		// Should log the error - this will fail until we implement structured logging
		logOutput := strings.TrimSpace(buf.String())
		if logOutput == "" {
			t.Fatal("Expected JSON encoding error to be logged")
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(logOutput), &logEntry); err != nil {
			t.Fatalf("Failed to parse log JSON: %v", err)
		}

		// Verify structured log fields
		if level := logEntry["level"]; level != "ERROR" {
			t.Errorf("Expected ERROR level, got: %v", level)
		}

		if msg, ok := logEntry["msg"]; !ok || !strings.Contains(msg.(string), "JSON") {
			t.Errorf("Expected JSON error message, got: %v", msg)
		}
	})
}
