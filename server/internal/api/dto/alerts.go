package dto

import "time"

type CreateVerifyEmail struct {
	Email   string `json:"email"`
	Project string `json:"project_name"`
}

type AlertMessage struct {
	MetricName    string    `json:"metric_name"`
	CurrentValue  float64   `json:"current_value"`
	Type          string    `json:"type"`
	ID            string    `json:"id"`
	ProjectName   string    `json:"project_name"`
	Operator      string    `json:"operator"`
	Threshold     string    `json:"threshold"`
	TimeWindow    string    `json:"time_window"`
	RuleType      string    `json:"rule_type"`
	LogField      string    `json:"log_field"`
	LogFieldValue string    `json:"log_field_value"`
	Methods       []Method  `json:"methods"`
	Timestamp     time.Time `json:"timestamp"`
	Priority      string    `json:"priority"`
	PublishedAt   time.Time `json:"published_at"`
	Source        string    `json:"source"`
	Version       string    `json:"version"`
}

type Method struct {
	Method string `json:"method"`
	Value  string `json:"value"`
}
