package redis_pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"server/internal/api/dto"
	serversentevents "server/internal/services/server_sent_events"
	"server/pkg"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/redis/go-redis/v9"
)

type AlertMonitor struct {
	Redis         *redis.Client
	Elasticsearch *elasticsearch.Client
	SSE           *serversentevents.SSEAlertService
}

func NewAlertMonitor(redis *redis.Client, elasticsearch *elasticsearch.Client, sse *serversentevents.SSEAlertService) *AlertMonitor {
	a := AlertMonitor{
		Redis:         redis,
		Elasticsearch: elasticsearch,
		SSE:           sse,
	}
	return &a
}

// StartMonitoring subscribes to a Redis pub/sub channel and processes alert messages
func (am *AlertMonitor) StartMonitoring(ctx context.Context) error {

	pubsub := am.Redis.Subscribe(ctx, "alerts")
	defer func(pubsub *redis.PubSub) {
		err := pubsub.Close()
		if err != nil {
		}
	}(pubsub)

	_, err := pubsub.Receive(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe to alerts channel: %w", err)
	}

	log.Println("Successfully subscribed to Redis alerts channel")

	// Start listening for messages
	ch := pubsub.Channel()
	for msg := range ch {
		if err := am.processAlert(ctx, msg.Payload); err != nil {
			log.Printf("Error processing alert: %v", err)
		}
	}

	return nil
}

// processAlert parses the alert message and saves it to Elasticsearch
func (am *AlertMonitor) processAlert(ctx context.Context, payload string) error {
	// Parse the JSON message
	var alert dto.AlertMessage
	if err := json.Unmarshal([]byte(payload), &alert); err != nil {
		return fmt.Errorf("failed to parse alert message: %w", err)
	}
	if client, ok := am.SSE.GetAlertClientChannel(alert.ProjectName); ok {
		log.Printf("Client channel found for service: %v", client)
	}
	if len(am.SSE.Clients) > 0 {
		am.SSE.BroadcastAlerts(alert.ProjectName, &payload)
	}
	indexName := fmt.Sprintf("alerts-%s-%s", strings.ToLower(strings.ReplaceAll(alert.ProjectName, " ", "_")), alert.Timestamp.Format("2006-01-02"))

	for _, method := range alert.Methods {
		if method.Method == "email" {
			subject, message := formatAlertMessage(alert)
			pkg.SendMail(method.Value, "alert", subject, message)
		}
	}

	// Save to Elasticsearch
	if err := am.saveToElasticsearch(ctx, indexName, alert); err != nil {
		return fmt.Errorf("failed to save to Elasticsearch: %w", err)
	}

	log.Printf("Successfully saved alert %s to index %s", alert.ID, indexName)
	return nil
}

// saveToElasticsearch indexes the alert document in Elasticsearch
func (am *AlertMonitor) saveToElasticsearch(ctx context.Context, indexName string, alert dto.AlertMessage) error {
	// Convert alert to JSON
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert to JSON: %w", err)
	}

	// Index the document
	res, err := am.Elasticsearch.Index(
		indexName,
		strings.NewReader(string(alertJSON)),
		am.Elasticsearch.Index.WithDocumentID(alert.ID),
		am.Elasticsearch.Index.WithContext(ctx),
		am.Elasticsearch.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Elasticsearch indexing error: %s", res.String())
	}

	return nil
}

func formatAlertMessage(alert dto.AlertMessage) (subject, message string) {
	// Common subject format
	subject = fmt.Sprintf("[%s Alert] %s in %s", strings.ToUpper(alert.Priority), alert.MetricName, alert.ProjectName)

	var body strings.Builder

	body.WriteString("<html><body style='font-family: Arial, sans-serif;'>")
	body.WriteString("<h2 style='color: #d9534f;'>Alert Notification</h2>")

	// Common alert information
	body.WriteString("<div style='margin-bottom: 15px;'>")
	body.WriteString(fmt.Sprintf("<p><strong>Project:</strong> %s</p>", alert.ProjectName))
	body.WriteString(fmt.Sprintf("<p><strong>Alert ID:</strong> %s</p>", alert.ID))
	body.WriteString(fmt.Sprintf("<p><strong>Priority:</strong> %s</p>", alert.Priority))
	body.WriteString(fmt.Sprintf("<p><strong>Triggered at:</strong> %s</p>", alert.Timestamp.Format(time.RFC1123)))
	body.WriteString("</div>")

	// Alert-specific details
	body.WriteString("<div style='background-color: #f8f9fa; padding: 15px; border-radius: 5px; margin-bottom: 15px;'>")
	body.WriteString("<h3 style='color: #337ab7;'>Alert Details</h3>")

	switch alert.Type {
	case "metric_avg":
		body.WriteString(fmt.Sprintf("<p>The average <strong>%s</strong> is currently <strong>%.2f</strong> which is %s the threshold of <strong>%s</strong>.</p>",
			alert.MetricName, alert.CurrentValue, getOperatorText(alert.Operator), alert.Threshold))

		// metric-specific guidance
		switch alert.MetricName {
		case "cpu_usage":
			body.WriteString("<p>This indicates that your system CPU usage has exceeded normal levels.</p>")
		case "memory_usage":
			body.WriteString("<p>This indicates that your system memory usage has exceeded normal levels.</p>")
		}

	case "log_count", "event_count":
		body.WriteString(fmt.Sprintf("<p>The count of logs with <strong>%s</strong> %s <strong>%s</strong> is currently <strong>%.0f</strong> which exceeds the threshold of <strong>%s</strong> within the time window of <strong>%s</strong>.</p>",
			alert.LogField, getOperatorText(alert.Operator), alert.Threshold, alert.CurrentValue, alert.Threshold, alert.TimeWindow))

		// log-specific guidance
		switch alert.LogField {
		case "status_code":
			body.WriteString("<p>This indicates an unusual number of HTTP status codes being returned.</p>")
		case "level":
			body.WriteString("<p>This indicates an unusual number of log messages at this severity level.</p>")
		case "ip_address":
			body.WriteString("<p>This indicates an unusual number of requests from a specific IP address.</p>")
		}
	}

	body.WriteString("</div>")

	body.WriteString("<div style='background-color: #e7f4ff; padding: 15px; border-radius: 5px;'>")
	body.WriteString("<h3 style='color: #337ab7;'>Recommended Actions</h3>")
	body.WriteString("<ul>")
	body.WriteString("<li>Review the metric/log details in your monitoring dashboard</li>")
	body.WriteString("<li>Check system health and recent deployments</li>")
	body.WriteString("<li>If this is unexpected, investigate potential issues</li>")
	body.WriteString("<li>Consider adjusting thresholds if alerts are too frequent</li>")
	body.WriteString("</ul>")
	body.WriteString("</div>")

	// Footer
	body.WriteString("<div style='margin-top: 20px; font-size: 12px; color: #777;'>")
	body.WriteString(fmt.Sprintf("<p>Alert generated by %s (v%s)</p>", alert.Source, alert.Version))
	body.WriteString("</div>")

	body.WriteString("</body></html>")

	return subject, body.String()
}

func getOperatorText(op string) string {
	switch op {
	case ">":
		return "above"
	case ">=":
		return "above or equal to"
	case "<":
		return "below"
	case "<=":
		return "below or equal to"
	case "==":
		return "equal to"
	case "!=":
		return "not equal to"
	default:
		return op
	}
}
