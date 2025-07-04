package config

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/riferrei/srclient"
	"google.golang.org/protobuf/proto"
)

type ProtobufSerializer struct {
	client *srclient.SchemaRegistryClient
}

// NewProtobufSerializer creates a new serializer instance.
func NewProtobufSerializer(client *srclient.SchemaRegistryClient) *ProtobufSerializer {
	return &ProtobufSerializer{client: client}
}

// Serialize encodes a protobuf message using the Confluent wire format.
// It prepends the message with a magic byte (0x0) and the 4-byte schema ID.
func (s *ProtobufSerializer) Serialize(topic string, message proto.Message) ([]byte, error) {

	if s == nil || s.client == nil {
		return nil, fmt.Errorf("ProtobufSerializer or its client is not initialized")
	}

	subject := topic + "-value"
	schema, err := s.client.GetLatestSchema(subject)
	if err != nil {

		log.Printf("Error: Failed to get schema for subject '%s'. %v", subject, err)
		return nil, fmt.Errorf("could not get schema for subject %s: %w", subject, err)
	}

	// Marshal the actual protobuf message to bytes.
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protobuf message: %w", err)
	}

	result := make([]byte, 0, 5+len(data))

	result = append(result, 0x0)

	schemaIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(schemaIDBytes, uint32(schema.ID()))
	result = append(result, schemaIDBytes...)

	result = append(result, data...)

	return result, nil
}

// SetupKafka initializes the Sarama SyncProducer and the ProtobufSerializer.
func SetupKafka(cfg *AppConfig) (sarama.SyncProducer, *ProtobufSerializer, error) {
	// Create Schema Registry client
	srClient := srclient.NewSchemaRegistryClient(cfg.SchemaRegistryURL)

	// Create a Protobuf Serializer
	protoSerializer := NewProtobufSerializer(srClient)

	// Sarama configuration
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack
	config.Producer.Retry.Max = 5                    // Retry up to 5 times
	config.Producer.Return.Successes = true          // Return successes on the success channel
	config.Producer.Return.Errors = true             // Return errors on the error channel

	// Create Sarama producer
	producer, err := sarama.NewSyncProducer([]string{cfg.KafkaHost}, config)
	if err != nil {
		return nil, nil, err
	}

	log.Println("Kafka producer created")
	return producer, protoSerializer, nil
}
