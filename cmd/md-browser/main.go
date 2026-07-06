package main

import (
	"log"

	"github.com/bvankampen/md-browser/internal/config"
	"github.com/bvankampen/md-browser/internal/server"
)

func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	srv := server.NewServer(cfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
