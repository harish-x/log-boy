package dto

type CpuUsagePoint struct {
	TimeLabel string  `json:"timeLabel"`
	Timestamp int64   `json:"timestamp"`
	Average   float64 `json:"average"`
}

type CpuUsageSeries struct {
	Label  string          `json:"label"`
	Points []CpuUsagePoint `json:"points"`
}

type MemoryUsagepoint struct {
	TimeLabel string  `json:"timeLabel"`
	Timestamp int64   `json:"timestamp"`
	Average   float64 `json:"average"`
}
