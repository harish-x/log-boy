package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"server/internal/api/dto"
	"server/pkg"

	"github.com/elastic/go-elasticsearch/v9"
)

type metricsES struct {
	es *elasticsearch.Client
}

func NewMetricsRepo(es *elasticsearch.Client) MetricsRepo {
	return &metricsES{es: es}
}

func (m *metricsES) GetCpuUsages(project string, from int64, to int64, groupBy string) ([]*dto.CpuUsagePoint, error) {

	var interval string
	switch groupBy {
	case "hour":
		interval = "1h"
	case "day":
		interval = "1d"
	default:
		return nil, fmt.Errorf("invalid groupBy value: must be 'hour' or 'day'")
	}

	lteValue := interface{}(to)
	if to == 0 {
		lteValue = "now"
	}

	// Elasticsearch query
	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"serviceName": project,
						},
					},
					map[string]interface{}{
						"range": map[string]interface{}{
							"cpuUsage.timestamp": map[string]interface{}{
								"gte": from,
								"lte": lteValue,
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"by_time": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":          "cpuUsage.timestamp",
					"fixed_interval": interval,
					"format":         "yyyy-MM-dd HH:mm",
					"time_zone":      "+05:30",
				},
				"aggs": map[string]interface{}{
					"avg_cpu": map[string]interface{}{
						"avg": map[string]interface{}{
							"field": "cpuUsage.average",
						},
					},
				},
			},
		},
	}

	// Marshal query to JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("encoding query: %w", err)
	}

	// Send the request
	res, err := m.es.Search(
		m.es.Search.WithContext(context.Background()),
		m.es.Search.WithIndex("m-"+project+"-*"),
		m.es.Search.WithBody(&buf),
		m.es.Search.WithTrackTotalHits(true),
		m.es.Search.WithPretty(),
	)
	if err != nil {
		return nil, fmt.Errorf("es search failed: %w", err)
	}
	defer res.Body.Close()

	// Parse response
	var esResp struct {
		Aggregations struct {
			ByTime struct {
				Buckets []struct {
					KeyAsString string `json:"key_as_string"`
					Key         int64  `json:"key"`
					AvgCPU      struct {
						Value float64 `json:"value"`
					} `json:"avg_cpu"`
				} `json:"buckets"`
			} `json:"by_time"`
		} `json:"aggregations"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	log.Print(esResp)
	var results []*dto.CpuUsagePoint
	for _, bucket := range esResp.Aggregations.ByTime.Buckets {
		if bucket.AvgCPU.Value == 0 {
			continue
		}

		results = append(results, &dto.CpuUsagePoint{
			TimeLabel: bucket.KeyAsString,
			Timestamp: bucket.Key,
			Average:   bucket.AvgCPU.Value,
		})
	}
	return results, nil
}

func (m *metricsES) GetMemoryUsages(project string, from int64, to int64, groupBy string) ([]*dto.MemoryUsagepoint, error) {

	var interval string
	switch groupBy {
	case "hour":
		interval = "1h"
	case "day":
		interval = "1d"
	default:
		return nil, fmt.Errorf("invalid groupBy value: must be 'hour' or 'day'")
	}

	// Elasticsearch query
	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"serviceName": project,
						},
					},
					map[string]interface{}{
						"range": map[string]interface{}{
							"memoryUsage.timestamp": map[string]interface{}{
								"gte": from,
								"lte": to,
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"by_time": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":          "memoryUsage.timestamp",
					"fixed_interval": interval,
					"format":         "yyyy-MM-dd HH:mm",
					"time_zone":      "+05:30",
				},
				"aggs": map[string]interface{}{
					"avg_cpu": map[string]interface{}{
						"avg": map[string]interface{}{
							"field": "memoryUsage.memoryUsagePercentage",
						},
					},
				},
			},
		},
	}

	// Marshal query to JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("encoding query: %w", err)
	}

	// Send the request
	res, err := m.es.Search(
		m.es.Search.WithContext(context.Background()),
		m.es.Search.WithIndex("m-"+project+"-*"),
		m.es.Search.WithBody(&buf),
		m.es.Search.WithTrackTotalHits(true),
		m.es.Search.WithPretty(),
	)
	if err != nil {
		return nil, fmt.Errorf("es search failed: %w", err)
	}
	defer res.Body.Close()

	// Parse response
	var esResp struct {
		Aggregations struct {
			ByTime struct {
				Buckets []struct {
					KeyAsString string `json:"key_as_string"`
					Key         int64  `json:"key"`
					AvgCPU      struct {
						Value float64 `json:"value"`
					} `json:"avg_cpu"`
				} `json:"buckets"`
			} `json:"by_time"`
		} `json:"aggregations"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	log.Print(esResp)
	var results []*dto.MemoryUsagepoint
	for _, bucket := range esResp.Aggregations.ByTime.Buckets {
		if bucket.AvgCPU.Value == 0 {
			continue
		}

		results = append(results, &dto.MemoryUsagepoint{
			TimeLabel: bucket.KeyAsString,
			Timestamp: bucket.Key,
			Average:   bucket.AvgCPU.Value,
		})
	}
	return results, nil
}

func (m *metricsES) GetMetricsMinMaxDate(project string) ([]*dto.MinMaxDate, error) {
	// query
	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"serviceName": project,
			},
		},
		"aggs": map[string]interface{}{
			"min_date": map[string]interface{}{
				"min": map[string]interface{}{
					"field": "memoryUsage.timestamp",
				},
			},
			"max_date": map[string]interface{}{
				"max": map[string]interface{}{
					"field": "memoryUsage.timestamp",
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("encoding query: %w", err)
	}

	res, err := m.es.Search(
		m.es.Search.WithContext(context.Background()),
		m.es.Search.WithIndex("m-"+project+"-*"),
		m.es.Search.WithBody(&buf),
		m.es.Search.WithTrackTotalHits(true),
		m.es.Search.WithPretty(),
	)
	if err != nil {
		return nil, fmt.Errorf("es search failed: %w", err)
	}
	defer res.Body.Close()

	var esResp struct {
		Aggregations struct {
			MinDate struct {
				Value float64 `json:"value"`
			} `json:"min_date"`
			MaxDate struct {
				Value float64 `json:"value"`
			} `json:"max_date"`
		} `json:"aggregations"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	log.Print(esResp)
	if esResp.Aggregations.MinDate.Value == 0 && esResp.Aggregations.MaxDate.Value == 0 {
		return nil, fmt.Errorf("no metrics found for the project")
	}

	if esResp.Aggregations.MinDate.Value == 0 || esResp.Aggregations.MaxDate.Value == 0 {
		return nil, fmt.Errorf("invalid date range: min or max date is zero")
	}

	return []*dto.MinMaxDate{
		{
			MinDate: pkg.ConvertEpochMillisToString(int64(esResp.Aggregations.MinDate.Value)),
			MaxDate: pkg.ConvertEpochMillisToString(int64(esResp.Aggregations.MaxDate.Value)),
		},
	}, nil
}
