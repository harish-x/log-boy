package services

import (
	"context"
	"log"
	"time"

	"github.com/IBM/sarama"

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
type DefaultLogProcessor struct{}

func NewDefaultLogProcessor() *DefaultLogProcessor {
	return &DefaultLogProcessor{}
}

func (p *DefaultLogProcessor) ProcessLog(logMessage *protogen.Log, topic string, partition int32, offset int64) error {
	log.Printf("Processing log - Topic: %s, Partition: %d, Offset: %d", topic, partition, offset)
	log.Printf("Log Details - Service: %s, Level: %s, Message: %s, Timestamp: %s",
		logMessage.GetServiceName(),
		logMessage.GetLevel(),
		logMessage.GetMessage(),
		logMessage.GetTimestamp().AsTime().Format(time.RFC3339))

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
