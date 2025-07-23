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
	PostgresDb        string
	GRPCSecret        string
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
		PostgresDb:        os.Getenv("POSTGRES_DB"),
		GRPCSecret:        os.Getenv("GRPC_SECRET"),
	}
	return &config, nil
}
