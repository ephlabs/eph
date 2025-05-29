package main

import (
	"log"

	"github.com/ephlabs/eph/internal/server"
)

func main() {
	log.Println("Starting Eph Daemon - What the eph?")

	if err := server.Run(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
