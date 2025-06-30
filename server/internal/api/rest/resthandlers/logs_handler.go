package resthandlers

import (
	"server/internal/repository"
	"server/internal/services"

	"github.com/gofiber/fiber/v2"
)

type LogsHandler struct {
	svc services.LogServices
}

func SetupLogsRoutes(r *RestHandler) {
	app := r.App
	svc := services.LogServices{
		Repo:   repository.NewLogRepo(r.ElasticSearch),
		Config: r.Config,
	}
	handler := LogsHandler{
		svc: svc,
	}
	api := app.Group("/api/v1/logs")
	api.Get("/:project", handler.GetLogs)
}

func (h LogsHandler) GetLogs(c *fiber.Ctx) error {
	project := c.Params("project")
	if project == "" {
		return BadRequestError(c, "Project name is required")
	}

	return nil
}
