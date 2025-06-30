package services

import (
	"server/config"
	"server/internal/api/dto"
	"server/internal/models"
	"server/internal/repository"
)

type LogServices struct {
	Repo   repository.LogRepo
	Config config.AppConfig
}

func (s *LogServices) GetLogs(filters *dto.LogFilter) ([]*models.Log, int64, error) {
	return s.Repo.GetLogs(filters)
}

func (s *LogServices) CreateProjectIndex(projectID string) error {
	return s.Repo.CreateProjectIndex(projectID)
}
