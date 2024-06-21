package main

import (
	"log"

	application "github.com/andymarkow/gophermart/internal/app"
)

func main() {
	app, err := application.New()
	if err != nil {
		log.Fatalf("application.New: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("application.Run: %v", err)
	}
}
