package config

import (
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	ServerPort              string
	ElasticSearchDNS        string
	ElasticSearchToken      string
	PostgresDb              string
	SynapseDb               string
	DirectoryTenantID       string
	ApplicationClientID     string
	AzureStorageAccountName string
	ColdStorageContainer    string
	AzureStorageKey         string
	KafkaBrokers            string
	SchemaRegistryURL       string
	ElasticsearchCloudID    string
}

func SetupEnv() (AppConfig, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return AppConfig{}, err
	}
	config := AppConfig{
		ServerPort:              os.Getenv("SERVER_PORT"),
		ElasticSearchDNS:        os.Getenv("ELASTIC_SEARCH_DNS"),
		PostgresDb:              os.Getenv("POSTGRES_DB"),
		SynapseDb:               os.Getenv("SYNAPSE_DB"),
		DirectoryTenantID:       os.Getenv("DIRECTORY_TENANT_ID"),
		ApplicationClientID:     os.Getenv("APPLICATION_CLIENT_ID"),
		AzureStorageAccountName: os.Getenv("AZURE_STORAGE_ACCOUNT_NAME"),
		ColdStorageContainer:    os.Getenv("COLD_STORAGE_CONTAINER"),
		AzureStorageKey:         os.Getenv("AZURE_STORAGE_KEY"),
		KafkaBrokers:            os.Getenv("KAFKA_BROKERS"),
		SchemaRegistryURL:       os.Getenv("SCHEMA_REGISTRY_URL"),
		ElasticSearchToken:      os.Getenv("ELASTIC_SEARCH_TOKEN"),
		ElasticsearchCloudID:    os.Getenv("ELASTICSEARCH_CLOUD_ID"),
	}
	return config, nil
}
