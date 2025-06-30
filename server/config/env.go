package config

import (
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	ServerPort              string
	ElasticSearch           string
	PostgresDb              string
	SynapseDb               string
	DirectoryTenantID       string
	ApplicationClientID     string
	AzureStorageAccountName string
	ColdStorageContainer    string
	AzureStorageKey         string
}

func SetupEnv() (AppConfig, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return AppConfig{}, err
	}
	config := AppConfig{
		ServerPort:              os.Getenv("SERVER_PORT"),
		ElasticSearch:           os.Getenv("ELASTIC_SEARCH"),
		PostgresDb:              os.Getenv("POSTGRES_DB"),
		SynapseDb:               os.Getenv("SYNAPSE_DB"),
		DirectoryTenantID:       os.Getenv("DIRECTORY_TENANT_ID"),
		ApplicationClientID:     os.Getenv("APPLICATION_CLIENT_ID"),
		AzureStorageAccountName: os.Getenv("AZURE_STORAGE_ACCOUNT_NAME"),
		ColdStorageContainer:    os.Getenv("COLD_STORAGE_CONTAINER"),
		AzureStorageKey:         os.Getenv("AZURE_STORAGE_KEY"),
	}
	return config, nil
}
