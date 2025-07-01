package resthandlers

import (
	"server/internal/api/dto"
	"server/internal/repository"
	"server/internal/services"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type LogsHandler struct {
	svc services.LogServices
}

func SetupLogsRoutes(r *RestHandler) {
	app := r.App
	svc := services.LogServices{
		Repo:   repository.NewLogRepo(r.ElasticSearch, r.SynapseDb),
		Config: r.Config,
	}
	handler := LogsHandler{
		svc: svc,
	}
	api := app.Group("/api/v1/logs")
	api.Get("/:project", handler.GetLogs)
	api.Get("/:project/date", handler.GetLogsMinMaxDates)
}

func (h LogsHandler) GetLogs(c *fiber.Ctx) error {
	project := c.Params("project")
	level := c.Query("level")
	limit, err := strconv.Atoi(c.Query("limit", "100"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid limit or page"})
	}
	offset, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid limit or page"})
	}
	fromStr := c.Query("from", "")
	toStr := c.Query("to", "")
	sortByDate := c.Query("sortByDate", "ASC")
	sortByDate = strings.ToUpper(sortByDate)
	var sortEnum = []string{"ASC", "DESC"}
	if ok := slices.Contains(sortEnum, sortByDate); !ok {
		sortByDate = "ASC"
	}
	var levelEnum = []string{"info", "debug", "warn", "error", "silly", "http", "verbose", ""}

	if ok := slices.Contains(levelEnum, level); !ok {
		return ErrorMessage(c, fiber.StatusBadRequest, "Invalid level")
	}

	if project == "" {
		return ErrorMessage(c, fiber.StatusBadRequest, "project name is required")
	}
	exists, err := h.svc.CheckIfIndexExists(project)
	if err != nil {
		return InternalError(c, err)
	}

	if exists == false {
		return ErrorMessage(c, fiber.StatusBadRequest, "project not found")
	}

	var fromFormatted, toFormatted string
	if fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return BadRequestError(c, "invalid from")
		}
		fromFormatted = from.Format(time.RFC3339)
	}

	if toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			return BadRequestError(c, "invalid to")
		}
		toFormatted = to.Format(time.RFC3339)
	}

	filter := &dto.LogFilter{
		Project:    project,
		Level:      level,
		Limit:      limit,
		Offset:     offset,
		From:       fromFormatted,
		To:         toFormatted,
		SortByDate: sortByDate,
	}
	logs, i, err := h.svc.GetLogs(filter)
	if err != nil {
		return err
	}
	return SuccessResponse(c, fiber.StatusOK, "Logs retrieved successfully", fiber.Map{"logs": logs, "total": i})
}

func (h LogsHandler) GetLogsMinMaxDates(c *fiber.Ctx) error {
	project := c.Params("project")
	if project == "" {
		return ErrorMessage(c, fiber.StatusBadRequest, "Project name is required")
	}
	exists, err := h.svc.CheckIfIndexExists(project)
	if err != nil {
		return InternalError(c, err)
	}
	if exists == false {
		return ErrorMessage(c, fiber.StatusBadRequest, "project not found")
	}
	dates, err := h.svc.GetLogsMinMaxDate(project)
	if err != nil {
		return ErrorMessage(c, fiber.StatusInternalServerError, err.Error())
	}
	return SuccessResponse(c, fiber.StatusOK, "Logs retrieved successfully", fiber.Map{
		"oldest": dates[0],
		"latest": dates[1],
	})
}
