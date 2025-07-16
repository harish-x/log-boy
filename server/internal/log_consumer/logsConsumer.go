package log_consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"server/internal/services"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/elastic/go-elasticsearch/v9"

	"server/config"
	"server/internal/models"
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

			// DeserializeLogs the message
			logMessage, err := h.deserializer.DeserializeLogs(message.Value)
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
	logSSE         *services.SSEService
}

type ServiceBatch struct {
	serviceName string
	buffer      []LogDocument
	flushTimer  *time.Timer
	mutex       sync.Mutex
}

type LogDocument struct {
	ServiceName    string        `json:"serviceName"`
	BuildDetails   *BuildDetails `json:"buildDetails,omitempty"`
	Level          string        `json:"level"`
	Message        string        `json:"message"`
	Stack          string        `json:"stack,omitempty"`
	RequestId      string        `json:"requestId,omitempty"`
	RequestUrl     string        `json:"requestUrl,omitempty"`
	RequestMethod  string        `json:"requestMethod,omitempty"`
	UserAgent      string        `json:"userAgent,omitempty"`
	RemoteIp       string        `json:"ipAddress,omitempty"`
	ResponseStatus string        `json:"responseStatus,omitempty"`
	ResponseTime   string        `json:"responseTime,omitempty"`
	Timestamp      time.Time     `json:"timestamp"`
	Topic          string        `json:"topic"`
	Partition      int32         `json:"partition"`
	Offset         int64         `json:"offset"`
}

type BuildDetails struct {
	NodeVersion string `json:"nodeVersion"`
	AppVersion  string `json:"appVersion"`
}

func NewDefaultLogProcessor(es *elasticsearch.Client, l *services.SSEService) *DefaultLogProcessor {
	return &DefaultLogProcessor{
		es:             es,
		serviceBatches: make(map[string]*ServiceBatch),
		batchSize:      100,
		flushInterval:  5 * time.Second,
		logSSE:         l,
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
		ServiceName:    serviceName,
		BuildDetails:   buildDetails,
		Level:          logMessage.GetLevel(),
		Message:        logMessage.GetMessage(),
		Stack:          logMessage.GetStack(),
		RequestId:      logMessage.GetRequestId(),
		RequestUrl:     logMessage.GetRequestUrl(),
		RequestMethod:  logMessage.GetRequestMethod(),
		UserAgent:      logMessage.GetUserAgent(),
		RemoteIp:       logMessage.GetRemoteIp(),
		Timestamp:      logMessage.GetTimestamp().AsTime(),
		ResponseStatus: logMessage.GetResponseStatus(),
		ResponseTime:   logMessage.GetResponseTime(),
		Topic:          topic,
		Partition:      partition,
		Offset:         offset,
	}
	if client, ok := p.logSSE.GetClientChannel(serviceName); ok {
		log.Printf("Client channel found for service: %v", client)
	}
	if len(p.logSSE.Clients) > 0 {
		logForBroadcast := toLogModel(doc)
		p.logSSE.BroadcastLogs(serviceName, logForBroadcast)
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	if res.IsError() {
		return fmt.Errorf("bulk request returned error for service %s: %s", sb.serviceName, res.String())
	}

	log.Printf("Successfully indexed %d logs documents for service: %s", len(sb.buffer), sb.serviceName)

	// Clear buffer
	sb.buffer = sb.buffer[:0]

	return nil
}

// FlushAll Force flush all service batches
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
}

func NewKafkaConsumerService(cfg *config.AppConfig, processor LogProcessor, consumerGroupID string) (*KafkaConsumerService, error) {
	consumerGroup, deserializer, err := config.SetupKafkaConsumer(cfg, consumerGroupID)
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

// This function converts from the database model to the broadcast model.
func toLogModel(doc LogDocument) *models.Log {
	var builddetails = models.BuildDetails{
		NodeVersion: doc.BuildDetails.NodeVersion,
		AppVersion:  doc.BuildDetails.AppVersion,
	}

	logModel := models.Log{
		ServiceName:    doc.ServiceName,
		Level:          doc.Level,
		Message:        doc.Message,
		Stack:          doc.Stack,
		RequestId:      doc.RequestId,
		RequestUrl:     doc.RequestUrl,
		RequestMethod:  doc.RequestMethod,
		UserAgent:      doc.UserAgent,
		Timestamp:      doc.Timestamp,
		IpAddress:      doc.RemoteIp,
		ResponseStatus: doc.ResponseStatus,
		ResponseTime:   doc.ResponseTime,
		BuildDetails:   builddetails,
	}

	return &logModel
}

// Stop gracefully shuts down the consumer
func (s *KafkaConsumerService) Stop() error {
	log.Println("Stopping Kafka consumer service...")
	return s.consumerManager.StopConsumer()
}
