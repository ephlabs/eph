package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ephlabs/eph/internal/log"
)

type Server struct {
	httpServer *http.Server
	config     *Config
	mu         sync.RWMutex
}

type Config struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		Port:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func New(cfg *Config) *Server {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Server{config: cfg}
}

func (s *Server) Start() error {
	mux := s.setupRoutes()

	server := &http.Server{
		Addr:         s.config.Port,
		Handler:      s.applyMiddleware(mux),
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	s.mu.Lock()
	s.httpServer = server
	s.mu.Unlock()

	log.Info(context.Background(), "Starting Eph daemon",
		"port", s.config.Port,
		"read_timeout", s.config.ReadTimeout,
		"write_timeout", s.config.WriteTimeout,
		"idle_timeout", s.config.IdleTimeout)
	return server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func Run() error {
	server := New(nil)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Info(context.Background(), "Shutting down gracefully", "timeout", 30*time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Error(context.Background(), "Server shutdown error", "error", err)
		}
	}()

	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}

func (s *Server) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error(context.Background(), "Failed to encode JSON response",
			"error", err,
			"status", status)
	}
}
