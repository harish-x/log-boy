package services

import (
	"errors"
	"server/internal/api/dto"
	"server/internal/repository"

	"gorm.io/gorm"
)

type MetricsServices struct {
	Repo       repository.MetricsRepo
	ProjetRepo repository.ProjectRepo
}

func (ms *MetricsServices) GetCpuUsage(project string, from int64, to int64, groupBy string) ([]*dto.CpuUsagePoint, error) {
	return ms.Repo.GetCpuUsages(project, from, to, groupBy)
}

func (ms *MetricsServices) GetMemoryUsage(project string, from int64, to int64, groupBy string) ([]*dto.MemoryUsagepoint, error) {
	return ms.Repo.GetMemoryUsages(project, from, to, groupBy)
}

func (ms *MetricsServices) CheckIfProjectExists(project string) (bool, error) {
	exists, err := ms.ProjetRepo.GetProjectByName(project)
	if err == gorm.ErrRecordNotFound {
		return false, errors.New("project not found")
	}
	if err != nil || exists == nil {
		return false, err
	}

	return true, nil

}


func(ms *MetricsServices) GetMetricsMinMaxDate(project string) ([]*dto.MinMaxDate, error) {
	dates, err := ms.Repo.GetMetricsMinMaxDate(project)
	if err != nil {
		return nil, err
	}
	if len(dates) == 0 {
		return nil, errors.New("no metrics found for the project")
	}
	return dates, nil

}