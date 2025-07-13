package repository

import (
	"server/internal/api/dto"
)

type MetricsRepo interface {
	GetCpuUsages(project string, from int64, to int64, groupBy string) ([]*dto.CpuUsagePoint, error)
	GetMemoryUsages(project string, from int64, to int64, groupBy string) ([]*dto.MemoryUsagepoint, error)
}
