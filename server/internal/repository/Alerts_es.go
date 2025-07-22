package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"server/internal/api/dto"
	"strings"
)

func (a *AlertRepo) GetAlerts(project string) (*[]dto.AlertMessage, error) {

	index := fmt.Sprintf("alerts-%s-*", project)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	queryBytes, err := json.Marshal(query)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	res, err := a.es.Search(a.es.Search.WithContext(context.Background()),
		a.es.Search.WithIndex(index),
		a.es.Search.WithBody(strings.NewReader(string(queryBytes))),
		a.es.Search.WithTrackTotalHits(true),
		a.es.Search.WithPretty())
	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(res.Body)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch returned error: %s", string(bodyBytes))
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
		return nil, fmt.Errorf("failed to unmarshal search result: %w", err)
	}

	var alerts []dto.AlertMessage

	for _, hit := range searchResult.Hits.Hits {
		sourceBytes, err := json.Marshal(hit.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal source: %w", err)
		}

		var alert dto.AlertMessage
		if err := json.Unmarshal(sourceBytes, &alert); err != nil {
			return nil, fmt.Errorf("failed to unmarshal alert message: %w", err)
		}

		alerts = append(alerts, alert)
	}

	return &alerts, nil
}
