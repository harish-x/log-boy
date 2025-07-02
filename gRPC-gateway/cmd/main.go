package main

import (
	"context"
	"gRPC-gateway/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errChan := make(chan error, 1)

	go func() {
		if err := server.StartNewgRPCServer(ctx); err != nil {
			errChan <- err
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Printf("Server error: %v", err)
	case sig := <-stopChan:
		log.Printf("Received signal: %v", sig)
	}

	log.Println("Shutting down...")
}
