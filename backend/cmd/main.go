package main

import (
	"log"
	"os"

	"pt-dana-sejahtera/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Printf("Application failed to start: %v", err)
		os.Exit(1)
	}
}