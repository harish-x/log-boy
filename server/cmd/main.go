package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"server/config"
	"server/internal/api/rest"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.SetupEnv()
	if err != nil {
		log.Fatal("Failed to load env variables", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)

	err = rest.StartRestServer(ctx, cfg)
	if err != nil {
		log.Printf("REST server error: %v", err)
		cancel()
		return
	}
	<-stopChan
	println("Shutting down servers...")
	cancel()
	time.Sleep(2 * time.Second)
	log.Println("Shutdown complete.")
}
