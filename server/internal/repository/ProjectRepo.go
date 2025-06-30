package repository

import (
	"server/internal/models"
)

type ProjectRepo interface {
	CreateProject(log *models.Project) (*models.Project, error)
	GetProjectByID(id string) (*models.Project, error)
	GetAllProjects(page int, limit int) ([]*models.Project, error)
	UpdateProject(project *models.Project) (*models.Project, error)
	DeleteProject(name string) error
	GetProjectByName(name string) (*models.Project, error)
	GetProjectsCount() (int64, error)
	GetLogs(projectName string) ([]*models.Log, error)
	GetRecentProjects(projectNames string) ([]*models.Project, error)
}
