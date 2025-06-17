package main

import (
	"context"
	"os"

	"github.com/ephlabs/eph/internal/log"
	"github.com/ephlabs/eph/internal/server"
)

func main() {
	logger := log.New()
	log.SetDefault(logger)

	log.Info(context.Background(), "Starting Eph Daemon - What the eph?")

	if err := server.Run(); err != nil {
		log.Error(context.Background(), "Server failed to start", "error", err)
		os.Exit(1)
	}
}
