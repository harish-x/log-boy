package redis_pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/redis/go-redis/v9"
)

type AlertMessage struct {
	MetricName   string    `json:"metric_name"`
	CurrentValue float64   `json:"current_value"`
	ID           string    `json:"id"`
	ProjectName  string    `json:"project_name"`
	Operator     string    `json:"operator"`
	Threshold    string    `json:"threshold"`
	TimeWindow   string    `json:"time_window"`
	RuleType     string    `json:"rule_type"`
	Methods      []Method  `json:"methods"`
	Timestamp    time.Time `json:"timestamp"`
	Priority     string    `json:"priority"`
	PublishedAt  time.Time `json:"published_at"`
	Source       string    `json:"source"`
	Version      string    `json:"version"`
}

type Method struct {
	Method string `json:"method"`
	Value  string `json:"value"`
}

type AlertMonitor struct {
	Redis         *redis.Client
	Elasticsearch *elasticsearch.Client
}

func NewAlertMonitor(redis *redis.Client, elasticsearch *elasticsearch.Client) *AlertMonitor {
	a := AlertMonitor{
		Redis:         redis,
		Elasticsearch: elasticsearch,
	}
	return &a
}

// StartMonitoring subscribes to a Redis pub/sub channel and processes alert messages
func (am *AlertMonitor) StartMonitoring(ctx context.Context) error {

	pubsub := am.Redis.Subscribe(ctx, "alerts")
	defer func(pubsub *redis.PubSub) {
		err := pubsub.Close()
		if err != nil {
		}
	}(pubsub)

	_, err := pubsub.Receive(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe to alerts channel: %w", err)
	}

	log.Println("Successfully subscribed to Redis alerts channel")

	// Start listening for messages
	ch := pubsub.Channel()
	for msg := range ch {
		if err := am.processAlert(ctx, msg.Payload); err != nil {
			log.Printf("Error processing alert: %v", err)
		}
	}

	return nil
}

// processAlert parses the alert message and saves it to Elasticsearch
func (am *AlertMonitor) processAlert(ctx context.Context, payload string) error {
	// Parse the JSON message
	var alert AlertMessage
	if err := json.Unmarshal([]byte(payload), &alert); err != nil {
		return fmt.Errorf("failed to parse alert message: %w", err)
	}

	indexName := fmt.Sprintf("alerts-%s-%s", alert.ProjectName, alert.Timestamp)
	// Save to Elasticsearch
	if err := am.saveToElasticsearch(ctx, indexName, alert); err != nil {
		return fmt.Errorf("failed to save to Elasticsearch: %w", err)
	}

	log.Printf("Successfully saved alert %s to index %s", alert.ID, indexName)
	return nil
}

// saveToElasticsearch indexes the alert document in Elasticsearch
func (am *AlertMonitor) saveToElasticsearch(ctx context.Context, indexName string, alert AlertMessage) error {
	// Convert alert to JSON
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert to JSON: %w", err)
	}

	// Index the document
	res, err := am.Elasticsearch.Index(
		indexName,
		strings.NewReader(string(alertJSON)),
		am.Elasticsearch.Index.WithDocumentID(alert.ID),
		am.Elasticsearch.Index.WithContext(ctx),
		am.Elasticsearch.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Elasticsearch indexing error: %s", res.String())
	}

	return nil
}

// StopMonitoring gracefully stops the monitoring process
func (am *AlertMonitor) StopMonitoring() {
	log.Println("Stopping alert monitoring...")
}

// Example usage
func main() {
	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // default DB
	})

	// Initialize Elasticsearch client
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}

	// Create AlertMonitor instance
	monitor := NewAlertMonitor(rdb, es)

	// Create context for graceful shutdown
	ctx := context.Background()

	// Start monitoring (this will block)
	log.Println("Starting alert monitoring...")
	if err := monitor.StartMonitoring(ctx); err != nil {
		log.Fatalf("Error in monitoring: %v", err)
	}
}
