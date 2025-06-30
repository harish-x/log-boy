package config

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
)

func NewElasticSearchDB(dns string) (*elasticsearch.Client, error) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},

		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}
	if _, err := client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}
	log.Print("elasticsearch connection established")

	return client, nil
}
