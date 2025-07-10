package metrics_consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"server/config"
	metricProto "server/internal/services/proto/metrics"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/elastic/go-elasticsearch/v9"
)

type MetricProcessor interface {
	ProcessMetrics(metricsMessage *metricProto.Metrics, topic string, partition int32, offset int64) error
}

type ConsumerGroupHandler struct {
	deserializer *config.ProtobufDeserializer
	processor    MetricProcessor
}

func NewConsumerGroupHandler(deserializer *config.ProtobufDeserializer, processor MetricProcessor) *ConsumerGroupHandler {
	return &ConsumerGroupHandler{
		deserializer: deserializer,
		processor:    processor,
	}
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Println("Consumer group session setup")
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Consumer group session cleanup")
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("Starting to consume from topic: %s, partition: %d", claim.Topic(), claim.Partition())

	// Process messages
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("Message channel closed for topic: %s, partition: %d", claim.Topic(), claim.Partition())
				return nil
			}

			if message == nil {
				continue
			}

			log.Printf("Received message from topic: %s, partition: %d, offset: %d",
				message.Topic, message.Partition, message.Offset)

			// DeserializeLogs the message
			metricMessage, err := h.deserializer.DeserializeMetrics(message.Value)
			if err != nil {
				log.Printf("Failed to deserialize message: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			// Process the log message
			err = h.processor.ProcessMetrics(metricMessage, message.Topic, message.Partition, message.Offset)
			if err != nil {
				log.Printf("Failed to process log message: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			session.MarkMessage(message, "")
			log.Printf("Successfully processed message from topic: %s, offset: %d",
				message.Topic, message.Offset)

		case <-session.Context().Done():
			log.Println("Consumer session context cancelled")
			return nil
		}
	}
}

// DefaultMetricsProcessor implements basic log processing
type DefaultMetricsProcessor struct {
	es             *elasticsearch.Client
	serviceBatches map[string]*ServiceBatch
	batchSize      int
	flushInterval  time.Duration
	mutex          sync.RWMutex
}

type ServiceBatch struct {
	serviceName string
	buffer      []Metrics
	flushTimer  *time.Timer
	mutex       sync.Mutex
}

func NewDefaultMetricsProcessor(es *elasticsearch.Client) *DefaultMetricsProcessor {
	return &DefaultMetricsProcessor{
		es:             es,
		serviceBatches: make(map[string]*ServiceBatch),
		batchSize:      100,
		flushInterval:  20 * time.Second,
	}
}

type MemoryUsage struct {
	Timestamp             int64   `json:"timestamp"`
	TotalMemory           int64   `json:"totalMemory"`
	FreeMemory            int64   `json:"freeMemory"`
	UsedMemory            int64   `json:"usedMemory"`
	MemoryUsagePercentage float64 `json:"memoryUsagePercentage"`
}
type CoreUsage struct {
	Core  int32   `json:"core"`
	Usage float64 `json:"usage"`
}
type CpuUsage struct {
	Timestamp int64        `json:"timestamp"`
	Average   float64      `json:"average"`
	Cores     []*CoreUsage `json:"cores"`
}
type Metrics struct {
	MemoryUsage *MemoryUsage `json:"memoryUsage"`
	CpuUsage    *CpuUsage    `json:"cpuUsage"`
	ServiceName string       `json:"serviceName"`
	Topic       string       `json:"topic"`
	Partition   int32        `json:"partition"`
	Offset      int64        `json:"offset"`
}

func (p *DefaultMetricsProcessor) ProcessMetrics(metrics *metricProto.Metrics, topic string, partition int32, offset int64) error {
	serviceName := metrics.GetServiceName()
	if serviceName == "" {
		return fmt.Errorf("service name is required")
	}

	batch := p.getOrCreateServiceBatch(serviceName)

	memoryUsage := MemoryUsage{
		Timestamp:             metrics.MemoryUsage.GetTimestamp(),
		TotalMemory:           metrics.MemoryUsage.GetTotalMemory(),
		FreeMemory:            metrics.MemoryUsage.GetFreeMemory(),
		UsedMemory:            metrics.MemoryUsage.GetUsedMemory(),
		MemoryUsagePercentage: metrics.MemoryUsage.GetMemoryUsagePercentage(),
	}
	var coreUsage []*CoreUsage

	for _, u := range metrics.GetCpuUsage().Cores {
		c := CoreUsage{
			Core:  u.GetCore(),
			Usage: u.GetUsage(),
		}
		coreUsage = append(coreUsage, &c)
	}
	cpuUsage := CpuUsage{
		Timestamp: metrics.CpuUsage.GetTimestamp(),
		Average:   metrics.CpuUsage.GetAverage(),
		Cores:     coreUsage,
	}
	metricsData := Metrics{
		MemoryUsage: &memoryUsage,
		CpuUsage:    &cpuUsage,
		ServiceName: serviceName,
		Topic:       topic,
		Partition:   partition,
		Offset:      offset,
	}
	return batch.addDocument(metricsData, p.batchSize, p.flushInterval, p.es)
}
func (p *DefaultMetricsProcessor) getOrCreateServiceBatch(serviceName string) *ServiceBatch {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	batch, exists := p.serviceBatches[serviceName]
	if !exists {
		batch = &ServiceBatch{
			serviceName: serviceName,
			buffer:      make([]Metrics, 0, p.batchSize),
		}
		p.serviceBatches[serviceName] = batch
	}

	return batch
}

func (sb *ServiceBatch) addDocument(doc Metrics, batchSize int, flushInterval time.Duration, client *elasticsearch.Client) error {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	// Add to buffer
	sb.buffer = append(sb.buffer, doc)

	// Reset flush timer
	if sb.flushTimer != nil {
		sb.flushTimer.Stop()
	}
	sb.flushTimer = time.AfterFunc(flushInterval, func() {
		err := sb.flushBatch(client)
		if err != nil {
			return
		}
	})

	// Flush if the batch is full
	if len(sb.buffer) >= batchSize {
		return sb.flushBatch(client)
	}

	return nil
}

func (sb *ServiceBatch) flushBatch(client *elasticsearch.Client) error {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	if len(sb.buffer) == 0 {
		return nil
	}

	var buf bytes.Buffer
	indexName := fmt.Sprintf("metrics-%s-%s", sb.serviceName, time.Now().Format("02.01.2006"))

	for _, doc := range sb.buffer {
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    fmt.Sprintf("%s-%d-%d", doc.Topic, doc.Partition, doc.Offset),
			},
		}
		actionBytes, _ := json.Marshal(action)
		buf.Write(actionBytes)
		buf.WriteByte('\n')

		// Document line
		docBytes, _ := json.Marshal(doc)
		buf.Write(docBytes)
		buf.WriteByte('\n')
	}

	// Send bulk request
	res, err := client.Bulk(
		bytes.NewReader(buf.Bytes()),
		client.Bulk.WithIndex(indexName),
		client.Bulk.WithRefresh("false"),
	)
	if err != nil {
		return fmt.Errorf("bulk request failed for service %s: %w", sb.serviceName, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	if res.IsError() {
		return fmt.Errorf("bulk request returned error for service %s: %s", sb.serviceName, res.String())
	}

	log.Printf("Successfully indexed %d documents for service: %s", len(sb.buffer), sb.serviceName)

	// Clear buffer
	sb.buffer = sb.buffer[:0]

	return nil
}

// FlushAll Force flush all service batches
func (p *DefaultMetricsProcessor) FlushAll() error {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var errors []string
	for serviceName, batch := range p.serviceBatches {
		if err := batch.flushBatch(p.es); err != nil {
			errors = append(errors, fmt.Sprintf("service %s: %v", serviceName, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("flush errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

func (p *DefaultMetricsProcessor) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var errors []string

	for serviceName, batch := range p.serviceBatches {
		batch.mutex.Lock()
		if batch.flushTimer != nil {
			batch.flushTimer.Stop()
		}
		batch.mutex.Unlock()

		if err := batch.flushBatch(p.es); err != nil {
			errors = append(errors, fmt.Sprintf("service %s: %v", serviceName, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("close errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// KafkaConsumerService manages the Kafka consumer service
type KafkaConsumerService struct {
	handler         *ConsumerGroupHandler
	consumerManager *config.KafkaConsumerManager
}

func NewKafkaConsumerService(cfg *config.AppConfig, processor MetricProcessor) (*KafkaConsumerService, error) {
	consumerGroup, deserializer, err := config.SetupKafkaConsumer(cfg, "metrics-consumer-group")
	if err != nil {
		return nil, err
	}

	consumerManager := config.NewKafkaConsumerManager(consumerGroup)

	handler := NewConsumerGroupHandler(deserializer, processor)

	return &KafkaConsumerService{
		handler:         handler,
		consumerManager: consumerManager,
	}, nil
}

// Start begins consuming messages from Kafka
func (s *KafkaConsumerService) Start(ctx context.Context, prefix string, ktm *config.KafkaTopicManager, refreshInterval time.Duration) error {
	errChan := make(chan error, 1)

	currentTopics, err := ktm.GetTopicsWithPrefix(prefix)
	if err != nil {
		return fmt.Errorf("initial topic fetch failed: %w", err)
	}

	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				topics, err := ktm.GetTopicsWithPrefix(prefix)
				if err != nil {
					log.Printf("Error fetching topics: %v", err)
					continue
				}

				if !equalStringSlices(currentTopics, topics) {
					log.Println("New topics detected, restarting consumer...")

					cancel()
					subCtx, cancel = context.WithCancel(ctx)

					currentTopics = topics
					validTopics := filterValidTopics(currentTopics, topics)
					if len(validTopics) == 0 {
						log.Println("No valid topics found. Skipping consumer restart.")
						continue
					}
					go func(topics []string) {
						err := s.consumerManager.StartConsumer(subCtx, topics, s.handler)
						if err != nil {
							log.Printf("Consumer restart error: %v", err)
						}
					}(validTopics)
				}

			case <-ctx.Done():
				log.Println("Shutting down topic watcher")
				cancel()
				return
			}
		}
	}()

	go func() {
		err := s.consumerManager.StartConsumer(subCtx, currentTopics, s.handler)
		if err != nil {
			errChan <- err
		}
		log.Println("Kafka consumer service started successfully")
	}()
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func filterValidTopics(input []string, validList []string) []string {
	validSet := make(map[string]struct{}, len(validList))
	for _, t := range validList {
		validSet[t] = struct{}{}
	}
	var result []string
	for _, t := range input {
		if _, ok := validSet[t]; ok {
			result = append(result, t)
		}
	}
	return result
}

// Stop gracefully shuts down the consumer
func (s *KafkaConsumerService) Stop() error {
	log.Println("Stopping Kafka consumer service...")
	return s.consumerManager.StopConsumer()
}
