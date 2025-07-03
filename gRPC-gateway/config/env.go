package config

import (
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	ServerPort        string
	KafkaHost         string
	PartitionStrategy string
	SchemaRegistryURL string
}

func SetupEnv() (*AppConfig, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	config := AppConfig{
		ServerPort:        os.Getenv("SERVER_PORT"),
		KafkaHost:         os.Getenv("KAFKA_HOST"),
		PartitionStrategy: os.Getenv("PARTITION_STRATEGY"),
		SchemaRegistryURL: os.Getenv("SCHEMA_REGISTRY_URL"),
	}
	return &config, nil
}
