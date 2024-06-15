package main

import (
	"log"

	"github.com/andymarkow/gophermart/internal/server"
)

func main() {
	srv, err := server.NewServer()
	if err != nil {
		log.Fatalf("server.NewServer: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("server.Start: %v", err)
	}
}
