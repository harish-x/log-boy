package config

import (
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	schemaregistry "github.com/djedjethai/gokfk-regent"
	"github.com/djedjethai/gokfk-regent/serde"
	gokfkprotobuf "github.com/djedjethai/gokfk-regent/serde/protobuf"
)

func SetupKafka(cfg *AppConfig) (*kafka.Producer, *gokfkprotobuf.Serializer, error) {

	srClient, err := schemaregistry.NewClient(schemaregistry.NewConfig(cfg.SchemaRegistryURL))
	if err != nil {
		return nil, nil, err
	}

	// Create a Protobuf Serializer using gokfk-regent
	protoSerializer, err := gokfkprotobuf.NewSerializer(srClient, serde.ValueSerde, gokfkprotobuf.NewSerializerConfig())
	if err != nil {
		return nil, nil, err
	}

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaHost,
	})
	if err != nil {
		return nil, nil, err
	}
	log.Println("Kafka producer created")
	return p, protoSerializer, nil
}
