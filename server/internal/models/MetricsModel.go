package models

type MemoryUsage struct {
	Timestamp             int64   `json:"timestamp"`
	TotalMemory           int64   `json:"totalMemory"`
	FreeMemory            int64   `json:"freeMemory"`
	UsedMemory            int64   `json:"usedMemory"`
	MemoryUsagePercentage float64 `json:"memoryUsagePercentage"`
}
type CoreUsage struct {
	Core  int32   `json:"core"`
	Usage float64 `json:"usage"`
}
type CpuUsage struct {
	Timestamp int64        `json:"timestamp"`
	Average   float64      `json:"average"`
	Cores     []*CoreUsage `json:"cores"`
}
type Metrics struct {
	MemoryUsage *MemoryUsage `json:"memoryUsage"`
	CpuUsage    *CpuUsage    `json:"cpuUsage"`
	ServiceName string       `json:"serviceName"`
}
