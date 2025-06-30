package repository

import (
	models2 "server/internal/models"
)

type ProjectRepo interface {
	CreateProject(log *models2.Project) (*models2.Project, error)
	GetProjectByID(id string) (*models2.Project, error)
	GetAllProjects(page int, limit int) ([]*models2.Project, error)
	UpdateProject(project *models2.Project) (*models2.Project, error)
	DeleteProject(name string) error
	GetProjectByName(name string) (*models2.Project, error)
	GetProjectsCount() (int64, error)
	GetLogs(projectName string) ([]*models2.Log, error)
	GetRecentProjects(projectNames string) ([]*models2.Project, error)
}
