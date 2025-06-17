package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ephlabs/eph/internal/log"
)

func TestLoggingMiddleware(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			if _, err := w.Write([]byte("test")); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
		})

		wrapped := loggingMiddleware(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})

	t.Run("structured logging with context", func(t *testing.T) {
		// Capture logs
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{}))
		log.SetDefault(logger)

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		})

		wrapped := loggingMiddleware(handler)

		// Create request with context that has request ID
		req := httptest.NewRequest("POST", "/api/environments", nil)
		ctx := log.WithRequestID(req.Context(), "test-123")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		// This test will fail until we implement structured logging
		logOutput := strings.TrimSpace(buf.String())
		if logOutput == "" {
			t.Fatal("Expected structured logs to be written")
		}

		// Parse the log entry
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(logOutput), &logEntry); err != nil {
			t.Fatalf("Failed to parse log JSON: %v", err)
		}

		// Verify structured log fields
		expectedFields := map[string]interface{}{
			"msg":        "HTTP request completed",
			"method":     "POST",
			"path":       "/api/environments",
			"status":     float64(201),
			"request_id": "test-123",
		}

		for key, expected := range expectedFields {
			if actual := logEntry[key]; actual != expected {
				t.Errorf("Expected %s=%v, got %v", key, expected, actual)
			}
		}

		// Verify duration field exists and is reasonable
		if duration, ok := logEntry["duration"]; !ok {
			t.Error("Expected duration field in log")
		} else if _, ok := duration.(float64); !ok {
			t.Errorf("Expected duration to be a number, got %T", duration)
		}
	})
}

func TestRequestIDMiddleware(t *testing.T) {
	// Capture context from handler
	var capturedCtx context.Context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	wrapped := requestIDMiddleware(handler)

	t.Run("generates request ID when not provided", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		if requestID == "" {
			t.Error("expected request ID header to be set")
		}

		// Verify it's a valid UUID format
		if len(requestID) != 36 || !strings.Contains(requestID, "-") {
			t.Errorf("expected valid UUID format, got %s", requestID)
		}

		// Test context integration - this will fail until we implement it
		if capturedCtx.Value(log.RequestIDKey) != requestID {
			t.Errorf("expected request ID in context to be %s, got %v", requestID, capturedCtx.Value(log.RequestIDKey))
		}
	})

	t.Run("uses provided request ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "test-request-id")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		if requestID != "test-request-id" {
			t.Errorf("expected X-Request-ID to be 'test-request-id', got %s", requestID)
		}

		// Test context integration
		if capturedCtx.Value(log.RequestIDKey) != "test-request-id" {
			t.Error("expected request ID in context to be 'test-request-id'")
		}
	})
}

func TestCORSMiddlewareAllMethods(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := corsMiddleware(handler)

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()

			wrapped.ServeHTTP(w, req)

			if w.Header().Get("Access-Control-Allow-Origin") != "*" {
				t.Error("expected CORS origin header to be '*'")
			}

			expectedMethods := strings.Join([]string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodDelete,
				http.MethodOptions,
			}, ", ")
			if w.Header().Get("Access-Control-Allow-Methods") != expectedMethods {
				t.Errorf("expected CORS methods header to be '%s', got '%s'", expectedMethods, w.Header().Get("Access-Control-Allow-Methods"))
			}

			expectedHeaders := "Content-Type, Authorization"
			if w.Header().Get("Access-Control-Allow-Headers") != expectedHeaders {
				t.Errorf("expected CORS headers header to be '%s'", expectedHeaders)
			}
		})
	}
}

func TestRecoveryMiddlewareNoPanic(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("no panic")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	wrapped := recoveryMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if !strings.Contains(w.Body.String(), "no panic") {
		t.Error("expected response body to contain 'no panic'")
	}
}

func TestResponseWriterWrite(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, status: 200}

	// Test Write method (should work normally)
	data := []byte("test data")
	n, err := rw.Write(data)
	if err != nil {
		t.Errorf("unexpected write error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected to write %d bytes, wrote %d", len(data), n)
	}
}

func TestMiddlewareOrder(t *testing.T) {
	// This test verifies that middleware are applied in the correct order
	server := New(nil)

	var executionOrder []string
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		executionOrder = append(executionOrder, "handler")
		w.WriteHeader(http.StatusOK)
	})

	// We can't directly test the order, but we can verify all middleware are applied
	wrapped := server.applyMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	// Verify all middleware effects are present
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("request ID middleware was not applied")
	}

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("CORS middleware was not applied")
	}

	if len(executionOrder) == 0 || executionOrder[len(executionOrder)-1] != "handler" {
		t.Error("handler was not executed")
	}
}
