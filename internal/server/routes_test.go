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
	req.SetPathValue("id", "test-env")
	w := httptest.NewRecorder()

	server.deleteEnvironment(w, req)

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
	req.SetPathValue("id", "test-env")
	w := httptest.NewRecorder()

	server.environmentLogs(w, req)

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
