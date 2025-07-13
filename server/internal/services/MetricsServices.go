package services

import (
	"server/internal/api/dto"
	"server/internal/repository"
)

type MetricsServices struct {
	Repo repository.MetricsRepo
}

func (ms *MetricsServices) GetCpuUsage(project string, from int64, to int64, groupBy string) ([]*dto.CpuUsagePoint, error) {
	return ms.Repo.GetCpuUsages(project, from, to, groupBy)
}

func (ms *MetricsServices) GetMemoryUsage(project string, from int64, to int64, groupBy string) ([]*dto.MemoryUsagepoint, error) {
	return ms.Repo.GetMemoryUsages(project, from, to, groupBy)
}
