//go:build integration

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestServerLifecycle(t *testing.T) {
	// Find a free port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	config := &Config{
		Port:         fmt.Sprintf(":%d", port),
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		IdleTimeout:  1 * time.Second,
	}

	server := New(config)

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Start()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test health endpoint
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify response content
	var healthResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}

	if healthResp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %s", healthResp["status"])
	}

	// Test graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("failed to shutdown server: %v", err)
	}

	// Verify server stopped
	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("unexpected server error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("server did not stop within timeout")
	}
}

func TestMultipleEndpoints(t *testing.T) {
	// Find a free port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	config := &Config{
		Port:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	server := New(config)

	// Start server
	go func() {
		server.Start()
	}()
	defer server.Shutdown(context.Background())

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	baseURL := fmt.Sprintf("http://localhost:%d", port)

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "health endpoint",
			path:           "/health",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]string
				json.Unmarshal(body, &resp)
				if resp["service"] != "ephd" {
					t.Errorf("expected service 'ephd', got %s", resp["service"])
				}
			},
		},
		{
			name:           "status endpoint",
			path:           "/api/v1/status",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				json.Unmarshal(body, &resp)
				if resp["status"] != "healthy" {
					t.Errorf("expected status 'healthy', got %v", resp["status"])
				}
			},
		},
		{
			name:           "list environments",
			path:           "/api/v1/environments",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				json.Unmarshal(body, &resp)
				if resp["total"].(float64) != 0 {
					t.Errorf("expected total 0, got %v", resp["total"])
				}
			},
		},
		{
			name:           "create environment",
			path:           "/api/v1/environments",
			method:         "POST",
			expectedStatus: http.StatusNotImplemented,
			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				json.Unmarshal(body, &resp)
				if resp["status"] != "not_implemented" {
					t.Errorf("expected status 'not_implemented', got %v", resp["status"])
				}
			},
		},
		{
			name:           "404 endpoint",
			path:           "/api/v1/nonexistent",
			method:         "GET",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]string
				json.Unmarshal(body, &resp)
				if resp["error"] != "Not found" {
					t.Errorf("expected error 'Not found', got %s", resp["error"])
				}
			},
		},
	}

	client := &http.Client{Timeout: 2 * time.Second}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, baseURL+tt.path, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Check headers
			if resp.Header.Get("X-Request-ID") == "" {
				t.Error("expected X-Request-ID header")
			}

			if resp.Header.Get("Access-Control-Allow-Origin") == "" {
				t.Error("expected CORS headers")
			}

			// Read body
			body := make([]byte, 0)
			buf := make([]byte, 1024)
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					body = append(body, buf[:n]...)
				}
				if err != nil {
					break
				}
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, body)
			}
		})
	}
}

func TestConcurrentRequests(t *testing.T) {
	// Find a free port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	config := &Config{
		Port:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	server := New(config)

	// Start server
	go func() {
		server.Start()
	}()
	defer server.Shutdown(context.Background())

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Make concurrent requests
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
			if err != nil {
				t.Errorf("request %d failed: %v", id, err)
			} else {
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("request %d: expected status 200, got %d", id, resp.StatusCode)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}
}

func TestServerShutdownTimeout(t *testing.T) {
	// Find a free port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	config := &Config{
		Port:         fmt.Sprintf(":%d", port),
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		IdleTimeout:  1 * time.Second,
	}

	server := New(config)

	// Start server
	go func() {
		server.Start()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test with already canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = server.Shutdown(ctx)
	if err == nil {
		t.Error("expected shutdown to fail with canceled context")
	}

	// Now shutdown properly
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	if err := server.Shutdown(ctx2); err != nil {
		t.Errorf("proper shutdown failed: %v", err)
	}
}

func TestRunFunction(t *testing.T) {
	// Test the full Run function with signal handling
	done := make(chan bool, 1)
	errChan := make(chan error, 1)

	// Start Run in a goroutine
	go func() {
		err := Run()
		errChan <- err
		done <- true
	}()

	// Wait for server to start
	time.Sleep(300 * time.Millisecond)

	// Send SIGTERM signal
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}
	if err := p.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send signal: %v", err)
	}

	// Wait for completion
	select {
	case <-done:
		err := <-errChan
		if err != nil {
			t.Errorf("Run() returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("Run() did not complete within timeout")
	}
}
