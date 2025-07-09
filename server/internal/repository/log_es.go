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
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"gorm.io/gorm"
)

type LogES struct {
	es      *elasticsearch.Client
	synapse *gorm.DB
}

type LogEntry struct {
	ServiceName   string    `json:"serviceName"`
	NodeVersion   string    `json:"nodeVersion"`
	AppVersion    string    `json:"appVersion"`
	Level         string    `json:"level"`
	Message       string    `json:"message"`
	RequestUrl    string    `json:"requestUrl"`
	RequestMethod string    `json:"requestMethod"`
	IpAddress     string    `json:"ipAddress"`
	UserAgent     string    `json:"userAgent"`
	RequestId     string    `json:"request_id"`
	Stack         string    `json:"stack"`
	Timestamp     time.Time `json:"timestamp"`
}

func (l *LogES) CreateProjectIndex(projectName string) error {
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
	if err != nil {
		return fmt.Errorf("failed to marshal index body: %w", err)
	}
	fmt.Printf("index body:\n %s \n", string(jsonBody))
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

		if errorType, ok := errorResponse["error"].(map[string]interface{})["type"].(string); ok {
			if errorType == "resource_already_exists_exception" {
				log.Print("Index was created by another process, which is fine")
				return nil
			}
		}

		return fmt.Errorf("elasticsearch error: %s", res.String())
	}

	log.Printf("Successfully created index: %s", indexName)
	return nil
}

func (l *LogES) GetLogs(filters *dto.LogFilter) ([]*models.Log, int64, error) {
	var logs []*models.Log

	if filters.Project == "" {
		return nil, 0, fmt.Errorf("project name is required")
	}

	// Build the Elasticsearch query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{},
			},
		},
		"sort": []map[string]interface{}{},
		"_source": []string{
			"service_name", "level", "message", "stack", "request_method",
			"request_url", "requestId", "remote_ip", "user_agent",
			"timestamp", "build_details",
		},
	}

	var mustQueries []map[string]interface{}

	// Filter by level
	if filters.Level != "" {
		mustQueries = append(mustQueries, map[string]interface{}{
			"term": map[string]interface{}{
				"level": filters.Level,
			},
		})
	}

	// Filter by date range
	if filters.From != "" || filters.To != "" {
		rangeQuery := map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{},
			},
		}

		if filters.From != "" {
			fromTime, err := time.Parse(time.RFC3339, filters.From)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid 'from' date format: %w", err)
			}
			rangeQuery["range"].(map[string]interface{})["timestamp"].(map[string]interface{})["gte"] = fromTime.Format(time.RFC3339)
		}

		if filters.To != "" {
			toTime, err := time.Parse(time.RFC3339, filters.To)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid 'to' date format: %w", err)
			}
			adjustedToTime := toTime.Add(24 * time.Hour)
			rangeQuery["range"].(map[string]interface{})["timestamp"].(map[string]interface{})["lt"] = adjustedToTime.Format(time.RFC3339)
		}

		mustQueries = append(mustQueries, rangeQuery)
	}

	// If no filters, use match_all
	if len(mustQueries) == 0 {
		query["query"] = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	} else {
		query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = mustQueries
	}

	// Add sorting
	if filters.SortByDate == "desc" {
		query["sort"] = []map[string]interface{}{
			{
				"timestamp": map[string]interface{}{
					"order": "desc",
				},
			},
		}
	} else {
		query["sort"] = []map[string]interface{}{
			{
				"timestamp": map[string]interface{}{
					"order": "asc",
				},
			},
		}
	}

	// Add pagination
	if filters.Offset > 0 && filters.Limit > 0 {
		query["from"] = (filters.Offset - 1) * filters.Limit
		query["size"] = filters.Limit
	} else if filters.Limit > 0 {
		query["size"] = filters.Limit
	}

	// Convert query to JSON
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal query: %w", err)
	}

	// Execute search
	res, err := l.es.Search(
		l.es.Search.WithContext(context.Background()),
		l.es.Search.WithIndex(filters.Project),
		l.es.Search.WithBody(strings.NewReader(string(queryBytes))),
		l.es.Search.WithTrackTotalHits(true),
		l.es.Search.WithPretty(),
	)

	if err != nil {
		return nil, 0, fmt.Errorf("elasticsearch search failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	// Read response
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response body: %w", err)
	}

	if res.IsError() {
		return nil, 0, fmt.Errorf("elasticsearch returned error: %s", string(bodyBytes))
	}

	// Parse response
	var searchResult struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.Unmarshal(bodyBytes, &searchResult); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal search result: %w", err)
	}

	// Convert hits to log models
	for _, hit := range searchResult.Hits.Hits {
		lg := &models.Log{}

		// Map fields from the source
		if serviceName, ok := hit.Source["service_name"].(string); ok {
			lg.ServiceName = serviceName
		}
		if level, ok := hit.Source["level"].(string); ok {
			lg.Level = level
		}
		if message, ok := hit.Source["message"].(string); ok {
			lg.Message = message
		}
		if stack, ok := hit.Source["stack"].(string); ok {
			lg.Stack = stack
		}
		if requestUrl, ok := hit.Source["request_url"].(string); ok {
			lg.RequestUrl = requestUrl
		}
		if requestMethod, ok := hit.Source["request_method"].(string); ok {
			lg.RequestMethod = requestMethod
		}
		if requestId, ok := hit.Source["requestId"].(string); ok {
			lg.RequestId = requestId
		}
		if ipAddress, ok := hit.Source["remote_ip"].(string); ok {
			lg.IpAddress = ipAddress
		}
		if userAgent, ok := hit.Source["user_agent"].(string); ok {
			lg.UserAgent = userAgent
		}

		if timestampStr, ok := hit.Source["timestamp"].(string); ok {
			if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
				lg.Timestamp = timestamp
			}
		}

		if buildDetailsRaw, ok := hit.Source["buildDetails"]; ok {
			if buildDetailsStr, ok := buildDetailsRaw.(string); ok && buildDetailsStr != "" {
				buildDetails := &models.BuildDetails{}
				if err := json.Unmarshal([]byte(buildDetailsStr), buildDetails); err == nil {
					lg.BuildDetails = *buildDetails
				}
			} else if buildDetailsMap, ok := buildDetailsRaw.(map[string]interface{}); ok {
				buildDetailsBytes, _ := json.Marshal(buildDetailsMap)
				buildDetails := &models.BuildDetails{}
				if err := json.Unmarshal(buildDetailsBytes, buildDetails); err == nil {
					lg.BuildDetails = *buildDetails
				}
			}
		}

		logs = append(logs, lg)
	}

	return logs, searchResult.Hits.Total.Value, nil

}

func (l *LogES) BulkInsertLogs(logs []*models.Log) error {
	//TODO implement me
	panic("implement me")
}

func (l *LogES) GetLogsAvailabilities(projectName string) ([]string, error) {
	var dates []string

	query := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"oldest_log": map[string]interface{}{
				"min": map[string]interface{}{
					"field": "timestamp",
				},
			},
			"latest_log": map[string]interface{}{
				"max": map[string]interface{}{
					"field": "timestamp",
				},
			},
		},
	}

	// Convert query to JSON
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// Execute search
	res, err := l.es.Search(
		l.es.Search.WithContext(context.Background()),
		l.es.Search.WithIndex(projectName),
		l.es.Search.WithBody(strings.NewReader(string(queryBytes))),
		l.es.Search.WithPretty(),
	)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch search failed: %w", err)
	}
	defer res.Body.Close()

	// Read response
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors in response
	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch returned error: %s", string(bodyBytes))
	}

	// Parse response
	var searchResult struct {
		Aggregations struct {
			OldestLog struct {
				Value         *float64 `json:"value"`
				ValueAsString *string  `json:"value_as_string"`
			} `json:"oldest_log"`
			LatestLog struct {
				Value         *float64 `json:"value"`
				ValueAsString *string  `json:"value_as_string"`
			} `json:"latest_log"`
		} `json:"aggregations"`
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
		} `json:"hits"`
	}

	if err := json.Unmarshal(bodyBytes, &searchResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search result: %w", err)
	}

	// Check if we have any documents in the index
	if searchResult.Hits.Total.Value == 0 {
		return nil, nil // No logs found, return nil like the original
	}

	// Check if aggregations returned valid values
	if searchResult.Aggregations.OldestLog.Value == nil || searchResult.Aggregations.LatestLog.Value == nil {
		return nil, nil
	}

	// Convert timestamps to time.Time and format as RFC3339
	oldestTime := time.Unix(0, int64(*searchResult.Aggregations.OldestLog.Value*1000000))
	latestTime := time.Unix(0, int64(*searchResult.Aggregations.LatestLog.Value*1000000))

	if searchResult.Aggregations.OldestLog.ValueAsString != nil && searchResult.Aggregations.LatestLog.ValueAsString != nil {
		oldestParsed, err := time.Parse(time.RFC3339, *searchResult.Aggregations.OldestLog.ValueAsString)
		if err == nil {
			oldestTime = oldestParsed
		}
		latestParsed, err := time.Parse(time.RFC3339, *searchResult.Aggregations.LatestLog.ValueAsString)
		if err == nil {
			latestTime = latestParsed
		}
	}

	dates = append(dates, oldestTime.Format(time.RFC3339))
	dates = append(dates, latestTime.Format(time.RFC3339))

	return dates, nil
}
func (l *LogES) GetLogsFromArchiveStorage(ProjectName string, fileName string, filters *dto.LogFilter) ([]*models.Log, int64, error) {

	var lg []LogEntry
	var totalCount int64
	offset := (filters.Offset - 1) * filters.Limit
	var whereClauses []string
	var args []interface{}

	// Level Filter
	if filters.Level != "" {
		whereClauses = append(whereClauses, "level = ?")
		args = append(args, filters.Level)
	}

	// Timestamp Filters
	if filters.From != "" {
		whereClauses = append(whereClauses, "timestamp >= ?")
		args = append(args, filters.From)
	}
	if filters.To != "" {

		whereClauses = append(whereClauses, "timestamp < ?")
		toTime, err := time.Parse(time.RFC3339, filters.To)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid 'to' date format: %w", err)
		}
		adjustedToTime := toTime.Add(24 * time.Hour)
		args = append(args, adjustedToTime)
	}

	// Default to a condition that is always true if no filters are applied
	whereCondition := "1=1"
	if len(whereClauses) > 0 {
		whereCondition = strings.Join(whereClauses, " AND ")
	}

	totalCountQuery := fmt.Sprintf(`SELECT COUNT(*)
FROM OPENROWSET(
    BULK '%s/%s',
    DATA_SOURCE = 'raw_ext_source',
    FORMAT = 'CSV',
    FIELDTERMINATOR = '0x0b',
    FIELDQUOTE = '0x0b'
) WITH (
    jsonContent varchar(MAX)
) AS [result]`, ProjectName, fileName)

	l.synapse.Raw(totalCountQuery).Scan(&totalCount)

	query := fmt.Sprintf(`WITH ParsedLogs AS (
		SELECT
			JSON_VALUE(jsonContent, '$.service_name') AS serviceName,
			JSON_VALUE(jsonContent, '$.build_details.nodeVersion') AS nodeVersion,
			JSON_VALUE(jsonContent, '$.build_details.appVersion') AS appVersion,
			JSON_VALUE(jsonContent, '$.level') AS level,
			JSON_VALUE(jsonContent, '$.message') AS message,
			JSON_VALUE(jsonContent, '$.request_id') AS requestId,
			JSON_VALUE(jsonContent, '$.request_url') AS requestUrl,
			JSON_VALUE(jsonContent, '$.request_method') AS requestMethod,
			JSON_VALUE(jsonContent, '$.stack') AS stack,
			JSON_VALUE(jsonContent, '$.ip_address') AS ipAddress,
			JSON_VALUE(jsonContent, '$.user_agent') AS userAgent,
			TRY_CONVERT(datetime2, JSON_VALUE(jsonContent, '$.timestamp')) AS timestamp
		FROM OPENROWSET(
			BULK '%s/%s',
			DATA_SOURCE = 'raw_ext_source',
			FORMAT = 'CSV',
			FIELDTERMINATOR = '0x0b',
			FIELDQUOTE = '0x0b'
		) WITH (
			jsonContent varchar(MAX)
		) AS [result]
	),
	FilteredLogs AS (
		SELECT *,
			   ROW_NUMBER() OVER (ORDER BY timestamp DESC) as RowNum
		FROM ParsedLogs
		WHERE %s
	)
	SELECT serviceName, level, message, nodeVersion, appVersion, requestUrl, requestId, requestMethod, stack, ipAddress, timestamp
	FROM FilteredLogs
	WHERE RowNum BETWEEN %d AND %d
	ORDER BY timestamp %s`, ProjectName, fileName, whereCondition, offset, filters.Limit, filters.SortByDate)
	err := l.synapse.Raw(query, args...).Scan(&lg).Error

	logs := make([]*models.Log, 0, len(lg))
	log.Print(lg)
	if err != nil {
		return nil, 0, err
	}
	for _, flatLog := range lg {
		finalLog := &models.Log{
			ServiceName:   flatLog.ServiceName,
			Level:         flatLog.Level,
			Message:       flatLog.Message,
			RequestId:     flatLog.RequestId,
			RequestUrl:    flatLog.RequestUrl,
			RequestMethod: flatLog.RequestMethod,
			Stack:         flatLog.Stack,
			IpAddress:     flatLog.IpAddress,
			UserAgent:     flatLog.UserAgent,
			Timestamp:     flatLog.Timestamp,
			// construct the nested BuildDetails struct
			BuildDetails: models.BuildDetails{
				NodeVersion: flatLog.NodeVersion,
				AppVersion:  flatLog.AppVersion,
			},
		}
		logs = append(logs, finalLog)
	}

	return logs, totalCount, nil
}

func (l *LogES) GetArchiveLogMinMaxDate(projectName string, fileName string) ([]string, error) {
	var dates []string

	sampleQuery := fmt.Sprintf(`SELECT MIN(TRY_CONVERT(datetime2, JSON_VALUE(jsonContent, '$.timestamp'))) as oldest, MAX(TRY_CONVERT(datetime2, JSON_VALUE(jsonContent, '$.timestamp'))) as latest FROM OPENROWSET(
		BULK '%s/%s',
		DATA_SOURCE = 'raw_ext_source',
		FORMAT = 'CSV',
		FIELDTERMINATOR = '0x0b',
		FIELDQUOTE = '0x0b'
	) WITH (
		jsonContent varchar(MAX)
	) AS [result]`, projectName, fileName)

	var result struct {
		Oldest *time.Time `gorm:"column:oldest"`
		Latest *time.Time `gorm:"column:latest"`
	}

	err := l.synapse.Raw(sampleQuery).Scan(&result).Error
	if err != nil {
		log.Printf("Query error: %v", err)
	}

	log.Printf("Oldest: %v, Latest: %v", result.Oldest, result.Latest)

	if err != nil {
		return nil, err
	}
	if result.Oldest == nil && result.Latest == nil {
		return nil, nil
	}
	// Append the formatted timestamps to the dates slice
	if result.Oldest != nil {
		dates = append(dates, result.Oldest.Format(time.RFC3339))
	}
	if result.Latest != nil {
		dates = append(dates, result.Latest.Format(time.RFC3339))
	}

	return dates, nil

}

func (l *LogES) CheckIfIndexExists(indexName string) (bool, error) {
	exists, err := l.es.Indices.Exists([]string{indexName})
	if err != nil {
		return false, err
	}
	if exists.StatusCode == 200 {
		return true, nil
	}
	return false, nil
}

func NewLogRepo(es *elasticsearch.Client, synapse *gorm.DB) LogRepo {
	return &LogES{es: es, synapse: synapse}
}
