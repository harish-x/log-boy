package resthandlers

import (
	"errors"
	"server/internal/api/dto"
	"server/internal/models"
	"server/internal/repository"
	"server/internal/services"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	svc services.ProjectServices
}

func SetupProjectRoutes(r *RestHandler) {
	app := r.App

	logSvc := services.LogServices{
		Repo:   repository.NewLogRepo(r.ElasticSearch, r.SynapseDb),
		Config: r.Config,
	}
	svc := services.ProjectServices{
		Repo:        repository.NewProjectRepo(r.PostgresDb),
		LogServices: &logSvc,
		Config:      r.Config,
		Ktm:         r.Ktm,
	}
	handler := ProjectHandler{
		svc: svc,
	}
	api := app.Group("/api/v1/projects")

	api.Post("/", handler.CreateProject)
	api.Get("/", handler.GetAllProjects)
	api.Get("/recent/projects", handler.GetRecentProjects)

	api.Get("/:name", handler.GetProjectByName)
	api.Put("/:name", handler.UpdateProject)
	api.Delete("/:name", handler.DeleteProject)
	api.Get("/:name/key", handler.GenerateProjectKey)
	api.Get("/:project/logs/stats", handler.GenerateLogStats)
}

// CreateProject handles the creation of a new project based on the provided request body.
// It validates the project data, checks for duplicate names, and returns an appropriate response.
func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	var project models.Project
	if err := c.BodyParser(&project); err != nil {
		return BadRequestError(c, err.Error())
	}

	if err := services.ValidateProject(&project); err != nil {
		return BadRequestError(c, err.Error())
	}

	existingProject, err := h.svc.GetProjectByName(project.Name)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return InternalError(c, err)
	}

	if existingProject != nil {
		return ErrorMessage(c, fiber.StatusConflict, "project with this name already exists")
	}
	createdProject, err := h.svc.CreateProject(&project)
	key := h.svc.GenerateProjectKey(project.Name)
	if err != nil {
		return InternalError(c, err)
	}
	response := dto.CreateProjectDto{
		Name: project.Name,
		Key:  key,
		ID:   createdProject.ID,
	}
	return SuccessResponse(c, fiber.StatusCreated, "Project created successfully", response)
}

// GetAllProjects retrieves all projects with optional pagination parameters (page and limit) from the query string.
func (h *ProjectHandler) GetAllProjects(c *fiber.Ctx) error {
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return BadRequestError(c, "Invalid page parameter")
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return BadRequestError(c, "Invalid limit parameter")
	}

	projects, err := h.svc.GetAllProjects(page, limit)
	if err != nil {
		return InternalError(c, err)
	}

	projectLength, err := h.svc.GetProjectsCount()
	if err != nil {
		return InternalError(c, err)
	}
	projectDtos := make([]dto.GetProjectsDto, 0, len(projects))
	for _, p := range projects {
		projectDtos = append(projectDtos, dto.GetProjectsDto{
			Name:        p.Name,
			ID:          p.ID,
			Description: p.Description,
			CreatedAT:   p.CreatedAt,
		})
	}

	return SuccessResponse(c, fiber.StatusOK, "Projects retrieved successfully", fiber.Map{"projects": projectDtos, "total": projectLength})
}

// DeleteProject deletes a project based on its name, validates input, and returns appropriate responses for errors or success.
func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return BadRequestError(c, "Project name is required")
	}

	project, err := h.svc.GetProjectByName(name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return InternalError(c, err)
	}
	if project == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrorMessage(c, fiber.StatusNotFound, "Project not found")
	}

	err = h.svc.DeleteProject(name)
	if err != nil {
		return InternalError(c, err)
	}
	return SuccessResponse(c, fiber.StatusOK, "Project deleted successfully", nil)
}

// GenerateLogStats calculates log statistics (errors, warnings, requests) for a project and updates the project's metrics.
func (h *ProjectHandler) GenerateLogStats(c *fiber.Ctx) error {
	project := c.Params("project")
	if project == "" {
		return BadRequestError(c, "Project name is required")
	}

	proj, err := h.svc.GetProjectByName(project)
	if err != nil {
		return ErrorMessage(c, fiber.StatusNotFound, "Project not found")
	}

	logs, err := h.svc.GetLogs(proj.Name)
	if err != nil {
		return ErrorMessage(c, fiber.StatusNotFound, "Project not found")
	}
	var totalErrors, totalWarnings, totalRequests int
	for _, log := range logs {
		switch log.Level {
		case "error":
			totalErrors++
		case "warn":
			totalWarnings++
		}
	}
	totalRequests = len(logs)
	errorRatio := float32(0)
	if totalErrors+totalWarnings+totalRequests > 0 {
		errorRatio = float32(totalErrors) / float32(totalErrors+totalWarnings+totalRequests)
	}
	if proj.ErrorRatio != errorRatio ||
		proj.TotalErrors != totalErrors ||
		proj.TotalWarnings != totalWarnings ||
		proj.TotalRequests != totalRequests {

		updatedProject, err := h.svc.UpdateProject(&models.Project{
			ID:            proj.ID,
			ErrorRatio:    errorRatio,
			TotalErrors:   totalErrors,
			TotalWarnings: totalWarnings,
			TotalRequests: totalRequests,
			Name:          proj.Name,
			UpdatedAt:     time.Now(),
		})
		if err != nil {
			return InternalError(c, err)
		}
		return SuccessResponse(c, fiber.StatusOK, "Log stats generated successfully", updatedProject)
	}
	return SuccessResponse(c, fiber.StatusOK, "Log stats already up to date", proj)

}

// GetProjectByID retrieves a project by its ID from the path parameter and returns it in the response.
// Returns a 400 error if ID is missing, a 404 error if the project is not found, or a 500 error for internal issues.
func (h *ProjectHandler) GetProjectByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return BadRequestError(c, "Project ID is required")
	}

	project, err := h.svc.GetProjectByID(id)
	if err != nil {
		return InternalError(c, err)
	}
	if project == nil {
		return ErrorMessage(c, fiber.StatusNotFound, "Project not found")
	}

	return SuccessResponse(c, fiber.StatusOK, "Project retrieved successfully", project)
}

// GetProjectByName retrieves a project by its name from the path parameter and returns it in the response.
// Returns a 400 error if the name is missing, a 404 error if the project is not found, or a 500 error for internal issues.
func (h *ProjectHandler) GetProjectByName(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return BadRequestError(c, "Project name is required")
	}

	project, err := h.svc.GetProjectByName(name)
	if project == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrorMessage(c, fiber.StatusNotFound, "Project not found")
	}

	if err != nil {
		return InternalError(c, err)
	}
	return SuccessResponse(c, fiber.StatusOK, "Project retrieved successfully", project)
}

// UpdateProject updates an existing project with new details provided in the request body and returns the updated project or an error.
func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {

	var project *dto.UpdateProjectDto
	projectName := c.Params("name")
	if projectName == "" {
		return BadRequestError(c, "Project name is required")
	}
	if err := c.BodyParser(&project); err != nil {
		return BadRequestError(c, err.Error())
	}

	updatedProject, err := h.svc.UpdateProject(&models.Project{
		Name:            projectName,
		Description:     project.Description,
		Active:          project.Active,
		RetentionPeriod: project.RetentionPeriod,
	})
	if err != nil {
		return InternalError(c, err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Project updated successfully", updatedProject)
}

// GenerateProjectKey generates a unique project key based on the provided project name.
// Returns a 400 error if the project name is missing, or a 200 success response with the generated key.
func (h *ProjectHandler) GenerateProjectKey(c *fiber.Ctx) error {
	projectName := c.Params("name")
	if projectName == "" {
		return BadRequestError(c, "Project name is required")
	}
	key := h.svc.GenerateProjectKey(projectName)
	return SuccessResponse(c, fiber.StatusOK, "Project key generated successfully", fiber.Map{"key": key})
}

// GetRecentProjects retrieves a list of recently accessed projects based on query parameters and returns them in the response.
// The method handles errors during retrieval and formats the project data into a standard response format.
func (h *ProjectHandler) GetRecentProjects(c *fiber.Ctx) error {
	projectNames := c.Query("p", "project_1")

	projects, err := h.svc.GetRecentProjects(projectNames)
	if err != nil {
		return InternalError(c, err)
	}
	var projectDtos []dto.RecentProjects
	for _, project := range projects {
		projectDtos = append(projectDtos, dto.RecentProjects{
			Name:        project.Name,
			Description: project.Description,
			ID:          project.ID,
			Active:      project.Active,
			CreatedAt:   project.CreatedAt,
		})
	}
	return SuccessResponse(c, fiber.StatusOK, "Projects retrieved successfully", projectDtos)
}
