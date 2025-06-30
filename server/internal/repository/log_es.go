package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"server/internal/api/dto"
	"server/internal/models"
	"strings"

	"github.com/elastic/go-elasticsearch/v9"
)

type LogES struct {
	es *elasticsearch.Client
}

func (l LogES) CreateProjectIndex(projectName string) error {
	indexName := fmt.Sprintf("logs-%s", projectName)
	exists, err := l.es.Indices.Exists([]string{indexName})

	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}

	if exists.StatusCode == 200 {
		err := fmt.Sprintf("Index %s already exists", indexName)
		return errors.New(err)
	}
	// index schema template
	indexBody := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 1,
			"index": map[string]interface{}{
				"refresh_interval": "1s",
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"serviceName": map[string]interface{}{
					"type": "keyword",
				},
				"buildDetails": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"nodeVersion": map[string]interface{}{
							"type": "keyword",
						},
						"appVersion": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"level": map[string]interface{}{
					"type": "keyword",
				},
				"message": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":         "keyword",
							"ignore_above": 256,
						},
					},
				},
				"stack": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
				"requestId": map[string]interface{}{
					"type": "keyword",
				},
				"requestUrl": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":         "keyword",
							"ignore_above": 256,
						},
					},
				},
				"requestMethod": map[string]interface{}{
					"type": "keyword",
				},
				"ipAddress": map[string]interface{}{
					"type": "ip",
				},
				"userAgent": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":         "keyword",
							"ignore_above": 256,
						},
					},
				},
				"timestamp": map[string]interface{}{
					"type":   "date",
					"format": "strict_date_optional_time||epoch_millis",
				},
			},
		},
	}

	jsonBody, err := json.Marshal(indexBody)
	log.Printf("index body:\n %s \n", string(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to marshal index body: %w", err)
	}

	// Create the index
	res, err := l.es.Indices.Create(
		indexName,
		l.es.Indices.Create.WithContext(context.Background()),
		l.es.Indices.Create.WithBody(strings.NewReader(string(jsonBody))),
		l.es.Indices.Create.WithPretty(),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	// Check response
	if res.IsError() {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&errorResponse); err != nil {
			return fmt.Errorf("failed to parse error response: %w", err)
		}

		// Check if it's a "resource_already_exists_exception" which is fine
		if errorType, ok := errorResponse["error"].(map[string]interface{})["type"].(string); ok {
			if errorType == "resource_already_exists_exception" {
				// Index was created by another process, which is fine
				log.Print("Index was created by another process, which is fine")
				return nil
			}
		}

		return fmt.Errorf("elasticsearch error: %s", res.String())
	}

	log.Printf("Successfully created index: %s", indexName)
	return nil
}

func (l LogES) GetLogs(filters *dto.LogFilter) ([]*models.Log, int64, error) {
	//TODO implement me
	panic("implement me")
}

func (l LogES) BulkInsertLogs(logs []*models.Log) error {
	//TODO implement me
	panic("implement me")
}

func (l LogES) GetProjectByName(name string) (*models.Project, error) {
	//TODO implement me
	panic("implement me")
}

func (l LogES) GetLogsAvailabilities(projectName string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (l LogES) GetLogsFromArchiveStorage(ProjectName string, fileName string, filters *dto.LogFilter) ([]*models.Log, int64, error) {
	//TODO implement me
	panic("implement me")
}

func (l LogES) GetArchiveLogMinMaxDate(projectName string, fileName string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func NewLogRepo(es *elasticsearch.Client) LogRepo {
	return &LogES{es: es}
}
