package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"server/config"
	"server/internal/api/rest"
	"server/internal/services"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	cfg, err := config.SetupEnv()
	if err != nil {
		log.Fatalf("Failed to load env variables: %v", err)
	}
	elasticSearch, err := config.NewElasticSearchDB(cfg.ElasticSearch)
	if err != nil {
		log.Fatalf("Failed to load env variables: %v", err)
	}

	errChan := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("REST server starting...")
		if err := rest.StartRestServer(ctx, cfg, elasticSearch); err != nil {
			errChan <- fmt.Errorf("REST server error: %w", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Kafka consumer starting...")

		processor := services.NewDefaultLogProcessor(elasticSearch)
		topics := []string{"project_1"}

		consumerService, err := services.NewKafkaConsumerService(&cfg, topics, processor)
		if err != nil {
			errChan <- fmt.Errorf("failed to create consumer service: %w", err)
			return
		}

		if err := consumerService.Start(ctx); err != nil {
			errChan <- fmt.Errorf("kafka consumer error: %w", err)
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Printf("Error received from a service: %v. Initiating shutdown.", err)
	case sig := <-stopChan:
		log.Printf("Signal %v received. Initiating shutdown.", sig)
	}

	log.Println("Shutting down all services...")
	cancel()

	shutdownComplete := make(chan struct{})
	go func() {
		wg.Wait()
		close(shutdownComplete)
	}()

	select {
	case <-shutdownComplete:
		log.Println("All services shut down gracefully.")
	case <-time.After(5 * time.Second):
		log.Println("Shutdown timed out. Forcing exit.")
	}
	log.Println("Application shutdown complete.")
}
