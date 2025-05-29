package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSetupRoutes(t *testing.T) {
	server := New(nil)
	mux := server.setupRoutes()

	if mux == nil {
		t.Fatal("expected mux to be non-nil")
	}

	// Test that routes are properly registered by making requests
	tests := []struct {
		path           string
		expectedStatus int
	}{
		{"/health", http.StatusOK},
		{"/api/v1/status", http.StatusOK},
		{"/api/v1/environments", http.StatusOK},
		{"/unknown", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("path %s: expected status %d, got %d", tt.path, tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHealthHandlerMethods(t *testing.T) {
	server := New(nil)

	methods := []string{"POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/health", nil)
			w := httptest.NewRecorder()

			server.healthHandler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d for %s, got %d", http.StatusMethodNotAllowed, method, w.Code)
			}
		})
	}
}

func TestStatusHandlerMethods(t *testing.T) {
	server := New(nil)

	methods := []string{"POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/status", nil)
			w := httptest.NewRecorder()

			server.statusHandler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d for %s, got %d", http.StatusMethodNotAllowed, method, w.Code)
			}
		})
	}
}

func TestListEnvironments(t *testing.T) {
	server := New(nil)

	req := httptest.NewRequest("GET", "/api/v1/environments", nil)
	w := httptest.NewRecorder()

	server.listEnvironments(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["total"].(float64) != 0 {
		t.Errorf("expected total 0, got %v", response["total"])
	}

	if !strings.Contains(response["message"].(string), "coming soon") {
		t.Error("expected message to indicate feature is coming soon")
	}
}

func TestCreateEnvironment(t *testing.T) {
	server := New(nil)

	req := httptest.NewRequest("POST", "/api/v1/environments", nil)
	w := httptest.NewRecorder()

	server.createEnvironment(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected status %d, got %d", http.StatusNotImplemented, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "not_implemented" {
		t.Errorf("expected status 'not_implemented', got %v", response["status"])
	}

	if !strings.Contains(response["message"].(string), "What the eph") {
		t.Error("expected message to contain 'What the eph'")
	}
}

func TestDeleteEnvironment(t *testing.T) {
	server := New(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/environments/test-env", nil)
	w := httptest.NewRecorder()

	server.deleteEnvironment(w, req, "test-env")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected status %d, got %d", http.StatusNotImplemented, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["environment_id"] != "test-env" {
		t.Errorf("expected environment_id 'test-env', got %v", response["environment_id"])
	}

	if response["status"] != "not_implemented" {
		t.Errorf("expected status 'not_implemented', got %v", response["status"])
	}
}

func TestEnvironmentLogs(t *testing.T) {
	server := New(nil)

	req := httptest.NewRequest("GET", "/api/v1/environments/test-env/logs", nil)
	w := httptest.NewRecorder()

	server.environmentLogs(w, req, "test-env")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected status %d, got %d", http.StatusNotImplemented, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["environment_id"] != "test-env" {
		t.Errorf("expected environment_id 'test-env', got %v", response["environment_id"])
	}

	logs, ok := response["logs"].([]interface{})
	if !ok || len(logs) != 0 {
		t.Error("expected empty logs array")
	}
}

func TestNotFoundHandlerMessage(t *testing.T) {
	server := New(nil)

	req := httptest.NewRequest("GET", "/api/v1/unknown", nil)
	w := httptest.NewRecorder()

	server.notFoundHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["path"] != "/api/v1/unknown" {
		t.Errorf("expected path '/api/v1/unknown', got %s", response["path"])
	}

	if !strings.Contains(response["message"], "What the eph") {
		t.Error("expected message to contain 'What the eph'")
	}
}

func TestJSONResponse(t *testing.T) {
	server := New(nil)

	testData := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	w := httptest.NewRecorder()
	server.jsonResponse(w, http.StatusCreated, testData)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %s", w.Header().Get("Content-Type"))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["key1"] != "value1" {
		t.Errorf("expected key1 'value1', got %v", response["key1"])
	}

	if response["key2"].(float64) != 123 {
		t.Errorf("expected key2 123, got %v", response["key2"])
	}

	if response["key3"] != true {
		t.Errorf("expected key3 true, got %v", response["key3"])
	}
}

func TestEnvironmentIDHandlerRouting(t *testing.T) {
	server := New(nil)

	tests := []struct {
		name   string
		path   string
		method string
		check  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "logs endpoint",
			path:   "/api/v1/environments/env123/logs",
			method: "GET",
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Code != http.StatusNotImplemented {
					t.Errorf("expected status %d, got %d", http.StatusNotImplemented, w.Code)
				}
				var resp map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if resp["environment_id"] != "env123" {
					t.Errorf("expected environment_id 'env123', got %v", resp["environment_id"])
				}
			},
		},
		{
			name:   "delete endpoint",
			path:   "/api/v1/environments/env456",
			method: "DELETE",
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Code != http.StatusNotImplemented {
					t.Errorf("expected status %d, got %d", http.StatusNotImplemented, w.Code)
				}
				var resp map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if resp["environment_id"] != "env456" {
					t.Errorf("expected environment_id 'env456', got %v", resp["environment_id"])
				}
			},
		},
		{
			name:   "invalid method",
			path:   "/api/v1/environments/env789",
			method: "POST",
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Code != http.StatusMethodNotAllowed {
					t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.URL.Path = tt.path // Ensure URL.Path is set correctly
			w := httptest.NewRecorder()

			server.environmentIDHandler(w, req)

			tt.check(t, w)
		})
	}
}
