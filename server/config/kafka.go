package config

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/riferrei/srclient"
	"google.golang.org/protobuf/proto"

	protogen "server/internal/services/proto/logs"
)

// ProtobufDeserializer handles deserialization of protobuf messages from Kafka
type ProtobufDeserializer struct {
	client *srclient.SchemaRegistryClient
}

func NewProtobufDeserializer(client *srclient.SchemaRegistryClient) *ProtobufDeserializer {
	return &ProtobufDeserializer{client: client}
}

// Deserialize decodes a protobuf message using the Confluent wire format
func (d *ProtobufDeserializer) Deserialize(data []byte) (*protogen.Log, error) {
	if d == nil || d.client == nil {
		return nil, fmt.Errorf("ProtobufDeserializer or its client is not initialized")
	}

	if len(data) < 5 {
		return nil, fmt.Errorf("message too short: expected at least 5 bytes, got %d", len(data))
	}

	// Check magic byte
	if data[0] != 0x0 {
		return nil, fmt.Errorf("invalid magic byte: expected 0x0, got 0x%x", data[0])
	}

	// Extract schema ID
	schemaID := binary.BigEndian.Uint32(data[1:5])

	// Verify schema exists (optional validation)
	_, err := d.client.GetSchema(int(schemaID))
	if err != nil {
		log.Printf("Warning: Could not validate schema ID %d: %v", schemaID, err)
	}

	// Extract protobuf data
	protoData := data[5:]

	// Unmarshal protobuf message
	logMessage := &protogen.Log{}
	err = proto.Unmarshal(protoData, logMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf message: %w", err)
	}

	return logMessage, nil
}

// SetupKafkaConsumer initializes the Sarama ConsumerGroup and the ProtobufDeserializer
func SetupKafkaConsumer(cfg *AppConfig) (sarama.ConsumerGroup, *ProtobufDeserializer, error) {
	// Create a Schema Registry client
	srClient := srclient.NewSchemaRegistryClient(cfg.SchemaRegistryURL)

	// Create a Protobuf Deserializer
	protoDeserializer := NewProtobufDeserializer(srClient)

	// Sarama consumer configuration
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Group.Rebalance.Retry.Max = 5
	config.Consumer.Group.Rebalance.Retry.Backoff = 5 * time.Second
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Consumer.Group.Session.Timeout = 30 * time.Second
	config.Consumer.Group.Rebalance.Timeout = 60 * time.Second
	config.Net.DialTimeout = 10 * time.Second
	config.Net.ReadTimeout = 10 * time.Second
	config.Net.WriteTimeout = 10 * time.Second
	config.Metadata.Retry.Max = 5
	config.Metadata.Retry.Backoff = 5 * time.Second
	config.Version = sarama.MaxVersion

	// Create a consumer group
	consumerGroup, err := sarama.NewConsumerGroup([]string{cfg.KafkaBrokers}, "log-consumer-group", config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	log.Println("Kafka consumer group created")
	return consumerGroup, protoDeserializer, nil
}

// KafkaConsumerManager manages the consumer group lifecycle
type KafkaConsumerManager struct {
	consumerGroup sarama.ConsumerGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewKafkaConsumerManager(consumerGroup sarama.ConsumerGroup) *KafkaConsumerManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &KafkaConsumerManager{
		consumerGroup: consumerGroup,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// StartConsumer starts the consumer group with given topics and handler
func (m *KafkaConsumerManager) StartConsumer(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	if len(topics) == 0 {
		log.Println("No topics provided to consumer. Skipping StartConsumer.")
		return nil
	}

	errChan := make(chan error, 1)
	go func() {
		for {
			if err := m.consumerGroup.Consume(ctx, topics, handler); err != nil {
				if errors.Is(err, sarama.ErrClosedClient) {
					log.Println("Sarama consumer closed gracefully.")
					return
				}
				log.Printf("Error from consumer: %v", err)
				errChan <- err
				return
			}

			if ctx.Err() != nil {
				log.Println("Consumer context canceled, exiting consume loop.")
				return
			}
		}
	}()

	go func() {
		for err := range m.consumerGroup.Errors() {
			log.Printf("Asynchronous consumer group error: %v", err)
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		log.Println("Context canceled, shutting down StartConsumer.")
		return nil
	}
}

func (m *KafkaConsumerManager) StopConsumer() error {
	log.Println("Stopping Kafka consumer...")

	m.cancel()

	if err := m.consumerGroup.Close(); err != nil {
		log.Printf("Error closing consumer group: %v", err)
		return err
	}

	log.Println("Kafka consumer stopped successfully")
	return nil
}
