package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"survival/internal/infrastructure/network/websocket"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := websocket.NewServer(ctx, "3033")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	log.Println("Starting server on port 3033...")
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("Server shut down successfully")
}
