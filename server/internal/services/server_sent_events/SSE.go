package serversentevents

type SSEService struct {
	LogSSE    *SSELogService
	MetricSSE *SSEMetricsService
	AlertSSE  *SSEAlertService
}

func NewSSEService() *SSEService {
	logSSE := NewSSELogsService()
	metricSSE := NewSSEMetricsService()
	alertSSE := NewSSEAlertService()
	return &SSEService{
		LogSSE:    logSSE,
		MetricSSE: metricSSE,
		AlertSSE:  alertSSE,
	}
}
