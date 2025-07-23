package resthandlers

import (
	"errors"
	"log"
	"server/internal/api/dto"
	"server/internal/models"
	"server/internal/repository"
	"server/internal/services"
	"server/pkg"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	svc services.ProjectServices
}

func SetupProjectRoutes(r *RestHandler) {
	app := r.App

	svc := services.ProjectServices{
		Repo:   repository.NewProjectRepo(r.PostgresDb),
		Config: r.Config,
		Ktm:    r.Ktm,
	}
	handler := ProjectHandler{
		svc: svc,
	}
	api := app.Group("/api/v1/projects")
	project := api.Use(pkg.AuthMiddleware())
	project.Post("/", handler.CreateProject)
	project.Get("/", handler.GetAllProjects)
	project.Get("/recent/projects", handler.GetRecentProjects)

	project.Get("/:name", handler.GetProjectByName)
	project.Put("/:name", handler.UpdateProject)
	project.Delete("/:name", handler.DeleteProject)
	project.Get("/:name/key", handler.GenerateProjectKey)
}

// CreateProject handles the creation of a new project based on the provided request body.
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
		return InternalError(c, errors.New("error while checking project"))
	}
	log.Print("line: 60", err)
	if existingProject != nil {
		return ErrorMessage(c, fiber.StatusConflict, "project with this name already exists")
	}
	createdProject, err := h.svc.CreateProject(&project)

	if err != nil {
		return InternalError(c, errors.New("error while creating project"))
	}
	key, err := h.svc.GenerateProjectKey(project.Name)
	log.Print("line: 70", err)
	if err != nil {
		return InternalError(c, errors.New("error while generating project key"))
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

// GetProjectByID retrieves a project by its ID from the path parameter and returns it in the response.
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
func (h *ProjectHandler) GenerateProjectKey(c *fiber.Ctx) error {
	projectName := c.Params("name")
	if projectName == "" {
		return BadRequestError(c, "Project name is required")
	}
	key, err := h.svc.GenerateProjectKey(projectName)
	if err != nil {
		return InternalError(c, err)
	}
	return SuccessResponse(c, fiber.StatusOK, "Project key generated successfully", fiber.Map{"key": key})
}

// GetRecentProjects retrieves a list of recently accessed projects based on query parameters and returns them in the response.
func (h *ProjectHandler) GetRecentProjects(c *fiber.Ctx) error {
	projectNames := c.Query("p", "project_1")

	if projectNames == "" {
		return ErrorMessage(c, fiber.StatusBadRequest, "project name is required")
	}

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
