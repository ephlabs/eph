package server

import (
	"net/http"
	"strings"

	"github.com/ephlabs/eph/pkg/version"
)

func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", s.healthHandler)

	// API v1 routes
	mux.HandleFunc("/api/v1/status", s.statusHandler)
	mux.HandleFunc("/api/v1/environments", s.environmentsHandler)

	// Handle environment ID paths
	mux.HandleFunc("/api/v1/environments/", s.environmentIDHandler)

	// 404 handler for unknown routes
	mux.HandleFunc("/", s.notFoundHandler)

	return mux
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]string{
		"status":  "ok",
		"service": "ephd",
		"version": version.GetVersion(),
	}

	s.jsonResponse(w, http.StatusOK, response)
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":  "healthy",
		"version": version.GetVersion(),
		"uptime":  "unknown", // TODO: track actual uptime
		"environments": map[string]int{
			"total":  0,
			"active": 0,
		},
	}

	s.jsonResponse(w, http.StatusOK, response)
}

func (s *Server) environmentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listEnvironments(w, r)
	case http.MethodPost:
		s.createEnvironment(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listEnvironments(w http.ResponseWriter, _ *http.Request) {
	response := map[string]interface{}{
		"environments": []string{},
		"total":        0,
		"message":      "Environment listing coming soon! When ready, this will show all your ephemeral environments.",
	}

	s.jsonResponse(w, http.StatusOK, response)
}

func (s *Server) createEnvironment(w http.ResponseWriter, _ *http.Request) {
	response := map[string]interface{}{
		"message": "Environment creation coming soon! What the eph are you going to build?",
		"status":  "not_implemented",
	}

	s.jsonResponse(w, http.StatusNotImplemented, response)
}

func (s *Server) environmentIDHandler(w http.ResponseWriter, r *http.Request) {
	// Extract environment ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/environments/")
	parts := strings.Split(path, "/")
	envID := parts[0]

	if len(parts) > 1 && parts[1] == "logs" {
		s.environmentLogs(w, r, envID)
		return
	}

	if r.Method == http.MethodDelete {
		s.deleteEnvironment(w, r, envID)
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

func (s *Server) deleteEnvironment(w http.ResponseWriter, _ *http.Request, envID string) {
	response := map[string]interface{}{
		"message":        "Environment deletion coming soon!",
		"environment_id": envID,
		"status":         "not_implemented",
	}

	s.jsonResponse(w, http.StatusNotImplemented, response)
}

func (s *Server) environmentLogs(w http.ResponseWriter, _ *http.Request, envID string) {
	response := map[string]interface{}{
		"message":        "Log streaming coming soon!",
		"environment_id": envID,
		"logs":           []string{},
	}

	s.jsonResponse(w, http.StatusNotImplemented, response)
}

func (s *Server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"error":   "Not found",
		"message": "The requested endpoint does not exist. What the eph are you looking for?",
		"path":    r.URL.Path,
	}

	s.jsonResponse(w, http.StatusNotFound, response)
}
