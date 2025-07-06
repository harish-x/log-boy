package repository

import (
	"server/internal/api/dto"
	"server/internal/models"
)

type LogRepo interface {
	GetLogs(filters *dto.LogFilter) ([]*models.Log, int64, error)
	BulkInsertLogs(logs []*models.Log) error
	GetLogsAvailabilities(projectName string) ([]string, error)
	GetLogsFromArchiveStorage(ProjectName string, fileName string, filters *dto.LogFilter) ([]*models.Log, int64, error)
	GetArchiveLogMinMaxDate(projectName string, fileName string) ([]string, error)
	CreateProjectIndex(projectID string) error
	CheckIfIndexExists(indexName string) (bool, error)
}
