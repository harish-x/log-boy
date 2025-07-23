package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"regexp"
	"server/config"
	"server/internal/models"
	"server/internal/repository"
	"time"

	"github.com/hashicorp/go-uuid"
)

type ProjectServices struct {
	Repo   repository.ProjectRepo
	Config config.AppConfig
	Ktm    *config.KafkaTopicManager
}

// CreateProject creates a new project in the repository and returns the created project or an error if creation fails.
func (p *ProjectServices) CreateProject(project *models.Project) (*models.Project, error) {

	project, err := p.Repo.CreateProject(project)
	if err != nil {
		return nil, err
	}
	err = p.Ktm.CreateProjectTopic(project.Name)
	if err != nil {
		return nil, err
	}
	return project, nil
}

// GetAllProjects retrieves a paginated list of projects based on the provided page number and limit. It returns the projects or an error.
func (p *ProjectServices) GetAllProjects(page int, limit int) ([]*models.Project, error) {
	projects, err := p.Repo.GetAllProjects(page, limit)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// ValidateProject validates the properties of a given project according to defined rules and returns an error if invalid.
func ValidateProject(project *models.Project) error {
	var namePattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`) // the project name must start with a lowercase letter or underscore, contain only lowercase letters, numbers, or underscores, and no special characters or spaces
	switch {
	case project.Name == "":
		return errors.New("project name is required")
	case len(project.Name) < 3:
		return errors.New("project name must be at least 3 characters long")
	case !namePattern.MatchString(project.Name):
		return errors.New("project name must start with a lowercase letter or underscore, contain only lowercase letters, numbers, or underscores, and no special characters or spaces")
	case project.RetentionPeriod == "":
		return errors.New("project retention period is required")
	}

	return nil
}

// GetProjectByID retrieves a project by its unique identifier from the repository and returns the project or an error.
func (p *ProjectServices) GetProjectByID(id string) (*models.Project, error) {
	project, err := p.Repo.GetProjectByID(id)
	if err != nil {
		return nil, err
	}
	return project, nil
}

// GetProjectByName retrieves a project from the repository by its name and returns the project or an error.
func (p *ProjectServices) GetProjectByName(name string) (*models.Project, error) {
	log.Printf(" name from service %s", name)
	project, err := p.Repo.GetProjectByName(name)
	if err != nil {
		return nil, err
	}
	return project, nil
}

// DeleteProject removes a project from the repository by its name and returns an error if the deletion fails.
func (p *ProjectServices) DeleteProject(name string) error {
	err := p.Repo.DeleteProject(name)
	if err != nil {
		return err
	}
	return nil
}

// GenerateProjectKey generates a unique HMAC-based project key using the project name and a secret from the configuration.
func (p *ProjectServices) GenerateProjectKey(projectName string) (string, error) {
	h := hmac.New(sha256.New, []byte(p.Config.GRPCSecret))
	generatedUUID, err := uuid.GenerateUUID()
	if err != nil {
		return "", err
	}
	timestamp := time.Now().Unix()
	payload := fmt.Sprintf("%s.%s.%d", projectName, generatedUUID, timestamp)
	key := models.KeyStore{
		Key:       projectName,
		Value:     generatedUUID,
		Timestamp: timestamp,
	}
	err = p.Repo.UpsertKeyStore(&key)
	if err != nil {
		return "", err
	}
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil)), nil
}

// GetProjectsCount retrieves the total count of projects for pagination in the repository and returns it along with an error if any occurs.
func (p *ProjectServices) GetProjectsCount() (int64, error) {
	return p.Repo.GetProjectsCount()
}

// GetLogs retrieves the logs associated with the specified project name. It returns a slice of logs or an error if any occurs.
func (p *ProjectServices) GetLogs(projectName string) ([]*models.Log, error) {
	return p.Repo.GetLogs(projectName)
}

// UpdateProject updates an existing project in the repository and returns the updated project or an error if the update fails.
func (p *ProjectServices) UpdateProject(project *models.Project) (*models.Project, error) {
	updatedProject, err := p.Repo.UpdateProject(project)
	if err != nil {
		return nil, err
	}
	return updatedProject, nil
}

// GetRecentProjects retrieves a list of recently accessed projects based on the provided project names and returns them or an error.
func (p *ProjectServices) GetRecentProjects(projectNames string) ([]*models.Project, error) {
	return p.Repo.GetRecentProjects(projectNames)
}
