package dto

import "time"

type CreateProjectDto struct {
	Name string `json:"name"`
	Key  string `json:"key"`
	ID   string `json:"id"`
}

type GetProjectsDto struct {
	Name        string    `json:"name"`
	ID          string    `json:"id"`
	Description string    `json:"description"`
	CreatedAT   time.Time `json:"created_at"`
}

type UpdateProjectDto struct {
	Description     string `json:"description"`
	Active          bool   `json:"active"`
	RetentionPeriod string `json:"retention_period"`
}

type RecentProjects struct {
	Name        string    `json:"name"`
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}
