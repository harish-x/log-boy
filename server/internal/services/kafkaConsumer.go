package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/elastic/go-elasticsearch/v9"

	"server/config"
	protogen "server/internal/services/proto/logs"
)

type LogProcessor interface {
	ProcessLog(logMessage *protogen.Log, topic string, partition int32, offset int64) error
}

type ConsumerGroupHandler struct {
	deserializer *config.ProtobufDeserializer
	processor    LogProcessor
}

func NewConsumerGroupHandler(deserializer *config.ProtobufDeserializer, processor LogProcessor) *ConsumerGroupHandler {
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

			// Deserialize the message
			logMessage, err := h.deserializer.Deserialize(message.Value)
			if err != nil {
				log.Printf("Failed to deserialize message: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			// Process the log message
			err = h.processor.ProcessLog(logMessage, message.Topic, message.Partition, message.Offset)
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

// DefaultLogProcessor implements basic log processing
type DefaultLogProcessor struct {
	es             *elasticsearch.Client
	serviceBatches map[string]*ServiceBatch
	batchSize      int
	flushInterval  time.Duration
	mutex          sync.RWMutex
}

type ServiceBatch struct {
	serviceName string
	buffer      []LogDocument
	flushTimer  *time.Timer
	mutex       sync.Mutex
}

type LogDocument struct {
	ServiceName   string        `json:"service_name"`
	BuildDetails  *BuildDetails `json:"build_details,omitempty"`
	Level         string        `json:"level"`
	Message       string        `json:"message"`
	Stack         string        `json:"stack,omitempty"`
	RequestId     string        `json:"request_id,omitempty"`
	RequestUrl    string        `json:"request_url,omitempty"`
	RequestMethod string        `json:"request_method,omitempty"`
	UserAgent     string        `json:"user_agent,omitempty"`
	RemoteIp      string        `json:"remote_ip,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
	Topic         string        `json:"topic"`
	Partition     int32         `json:"partition"`
	Offset        int64         `json:"offset"`
}

type BuildDetails struct {
	NodeVersion string `json:"nodeVersion"`
	AppVersion  string `json:"appVersion"`
}

func NewDefaultLogProcessor(es *elasticsearch.Client) *DefaultLogProcessor {
	return &DefaultLogProcessor{
		es:             es,
		serviceBatches: make(map[string]*ServiceBatch),
		batchSize:      100,
		flushInterval:  5 * time.Second,
	}
}

func (p *DefaultLogProcessor) ProcessLog(logMessage *protogen.Log, topic string, partition int32, offset int64) error {
	serviceName := logMessage.GetServiceName()
	if serviceName == "" {
		return fmt.Errorf("service name is required")
	}

	batch := p.getOrCreateServiceBatch(serviceName)

	var buildDetails *BuildDetails
	if logMessage.GetBuildDetails() != nil {
		buildDetails = &BuildDetails{
			NodeVersion: logMessage.GetBuildDetails().GetAppVersion(),
			AppVersion:  logMessage.GetBuildDetails().GetAppVersion(),
		}
	}

	// Create log document
	doc := LogDocument{
		ServiceName:   serviceName,
		BuildDetails:  buildDetails,
		Level:         logMessage.GetLevel(),
		Message:       logMessage.GetMessage(),
		Stack:         logMessage.GetStack(),
		RequestId:     logMessage.GetRequestId(),
		RequestUrl:    logMessage.GetRequestUrl(),
		RequestMethod: logMessage.GetRequestMethod(),
		UserAgent:     logMessage.GetUserAgent(),
		RemoteIp:      logMessage.GetRemoteIp(),
		Timestamp:     logMessage.GetTimestamp().AsTime(),
		Topic:         topic,
		Partition:     partition,
		Offset:        offset,
	}

	return batch.addDocument(doc, p.batchSize, p.flushInterval, p.es)
}
func (p *DefaultLogProcessor) getOrCreateServiceBatch(serviceName string) *ServiceBatch {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	batch, exists := p.serviceBatches[serviceName]
	if !exists {
		batch = &ServiceBatch{
			serviceName: serviceName,
			buffer:      make([]LogDocument, 0, p.batchSize),
		}
		p.serviceBatches[serviceName] = batch
	}

	return batch
}

func (sb *ServiceBatch) addDocument(doc LogDocument, batchSize int, flushInterval time.Duration, client *elasticsearch.Client) error {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	// Add to buffer
	sb.buffer = append(sb.buffer, doc)

	// Reset flush timer
	if sb.flushTimer != nil {
		sb.flushTimer.Stop()
	}
	sb.flushTimer = time.AfterFunc(flushInterval, func() {
		sb.flushBatch(client)
	})

	// Flush if batch is full
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
	indexName := fmt.Sprintf("logs-%s", sb.serviceName)

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
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk request returned error for service %s: %s", sb.serviceName, res.String())
	}

	log.Printf("Successfully indexed %d documents for service: %s", len(sb.buffer), sb.serviceName)

	// Clear buffer
	sb.buffer = sb.buffer[:0]

	return nil
}

// Force flush all service batches
func (p *DefaultLogProcessor) FlushAll() error {
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

func (p *DefaultLogProcessor) Close() error {
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
	topics          []string
}

func NewKafkaConsumerService(cfg *config.AppConfig, topics []string, processor LogProcessor) (*KafkaConsumerService, error) {
	consumerGroup, deserializer, err := config.SetupKafkaConsumer(cfg)
	if err != nil {
		return nil, err
	}

	consumerManager := config.NewKafkaConsumerManager(consumerGroup)

	handler := NewConsumerGroupHandler(deserializer, processor)

	return &KafkaConsumerService{
		handler:         handler,
		consumerManager: consumerManager,
		topics:          topics,
	}, nil
}

// Start begins consuming messages from Kafka
func (s *KafkaConsumerService) Start(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		err := s.consumerManager.StartConsumer(ctx, s.topics, s.handler)
		if err != nil {
			errChan <- err
		}
		close(errChan)
		log.Println("Kafka consumer service started successfully")
	}()
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}

// Stop gracefully shuts down the consumer
func (s *KafkaConsumerService) Stop() error {
	log.Println("Stopping Kafka consumer service...")
	return s.consumerManager.StopConsumer()
}
