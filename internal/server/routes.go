package server

import (
	"net/http"

	"github.com/ephlabs/eph/pkg/version"
)

func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.healthHandler)
	mux.HandleFunc("GET /api/v1/status", s.statusHandler)
	mux.HandleFunc("GET /api/v1/environments", s.listEnvironments)
	mux.HandleFunc("POST /api/v1/environments", s.createEnvironment)
	mux.HandleFunc("DELETE /api/v1/environments/{id}", s.deleteEnvironment)
	mux.HandleFunc("GET /api/v1/environments/{id}/logs", s.environmentLogs)
	mux.HandleFunc("/", s.notFoundHandler)

	return mux
}

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	response := map[string]string{
		"status":  "ok",
		"service": "ephd",
		"version": version.GetVersion(),
	}

	s.jsonResponse(w, http.StatusOK, response)
}

func (s *Server) statusHandler(w http.ResponseWriter, _ *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"version": version.GetVersion(),
		"uptime":  "unknown",
		"environments": map[string]int{
			"total":  0,
			"active": 0,
		},
	}

	s.jsonResponse(w, http.StatusOK, response)
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

func (s *Server) deleteEnvironment(w http.ResponseWriter, r *http.Request) {
	envID := r.PathValue("id")
	response := map[string]interface{}{
		"message":        "Environment deletion coming soon!",
		"environment_id": envID,
		"status":         "not_implemented",
	}

	s.jsonResponse(w, http.StatusNotImplemented, response)
}

func (s *Server) environmentLogs(w http.ResponseWriter, r *http.Request) {
	envID := r.PathValue("id")
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
