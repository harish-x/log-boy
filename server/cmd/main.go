package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"server/config"
	"server/internal/api/rest"
	"server/internal/log_consumer"
	"server/internal/metrics_consumer"
	"server/internal/services"
	"strings"
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
	elasticSearch, err := config.NewElasticSearchDB(cfg.ElasticSearchDNS)
	if err != nil {
		log.Fatalf("Failed to load env variables: %v", err)
	}
	brokers := strings.Split(cfg.KafkaBrokers, ",")

	errChan := make(chan error, 2)
	ktm, err := config.NewKafkaTopicManager(brokers)
	defer func(ktm *config.KafkaTopicManager) {
		err := ktm.Close()
		if err != nil {
			errChan <- fmt.Errorf("failed to close kafka topic manager : %v", err)
		}
	}(ktm)
	if err != nil {
		errChan <- fmt.Errorf("failed to Start kafka topic manager : %v", err)
	}
	logSSE := services.NewSSEService()
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("REST server starting...")
		if err := rest.StartRestServer(ctx, cfg, elasticSearch, ktm, logSSE); err != nil {
			errChan <- fmt.Errorf("REST server error: %w", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Kafka consumer starting...")

		processor := log_consumer.NewDefaultLogProcessor(elasticSearch, logSSE)
		consumerGroupID := "log-consumer-group"
		consumerService, err := log_consumer.NewKafkaConsumerService(&cfg, processor, consumerGroupID)
		if err != nil {
			errChan <- fmt.Errorf("failed to create consumer service: %w", err)
			return
		}

		if err := consumerService.Start(ctx, "logs-", ktm, time.Minute*2); err != nil {
			errChan <- fmt.Errorf("kafka logs consumer error: %w", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		processor := metrics_consumer.NewDefaultMetricsProcessor(elasticSearch)
		consumerGroupId := "metrics-consumer-group"
		consumerService, err := metrics_consumer.NewKafkaConsumerService(&cfg, processor, consumerGroupId)
		if err != nil {
			errChan <- fmt.Errorf("failed to create consumer service: %w", err)
			return
		}

		if err := consumerService.Start(ctx, "metrics-", ktm, time.Minute*2); err != nil {
			errChan <- fmt.Errorf("kafka metrics consumer error: %w", err)
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Printf("-Error received from a service: %v. Initiating shutdown.", err)
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
