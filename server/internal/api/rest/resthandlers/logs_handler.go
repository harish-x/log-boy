package resthandlers

import (
	"bufio"
	"encoding/json"
	"log"
	"server/internal/api/dto"
	"server/internal/repository"
	"server/internal/services"
	"server/internal/services/server_sent_events"
	"server/pkg"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type LogsHandler struct {
	svc services.LogServices
	sse *serversentevents.SSELogService
}

func SetupLogsRoutes(r *RestHandler, l *serversentevents.SSELogService) {
	app := r.App
	svc := services.LogServices{
		Repo:   repository.NewLogRepo(r.ElasticSearch, r.SynapseDb),
		Config: r.Config,
	}
	handler := LogsHandler{
		svc: svc,
		sse: l,
	}
	api := app.Group("/api/v1/logs")
	api.Get("/:project", pkg.AuthMiddleware(), handler.GetLogs)
	api.Get("/:project/date", pkg.AuthMiddleware(), handler.GetLogsMinMaxDates)
	api.Get("/:project/archives", pkg.AuthMiddleware(), handler.ListLogsFromArchive)
	api.Get("/:project/archive", pkg.AuthMiddleware(), handler.GetLogsFromColdStorage)
	api.Get("/:project/stream", pkg.SSEAuthMiddleware(), handler.StreamLogs)
}

func (h *LogsHandler) GetLogs(c *fiber.Ctx) error {
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
	sortByDate := c.Query("sortByDate", "asc")
	sortByDate = strings.ToLower(sortByDate)
	var sortEnum = []string{"asc", "desc"}
	if ok := slices.Contains(sortEnum, sortByDate); !ok {
		sortByDate = "asc"
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

	if !exists {
		return ErrorMessage(c, fiber.StatusBadRequest, "project not found or No logs found")
	}
	var fromFormatted, toFormatted string

	if fromStr != "" {
		fromFormatted, err = pkg.ParseFormettedTimeString(fromStr)
		if err != nil {
			return ErrorMessage(c, fiber.StatusBadRequest, err.Error())
		}
	}

	if toStr != "" {
		toFormatted, err = pkg.ParseFormettedTimeString(fromStr)
		if err != nil {
			return BadRequestError(c, err.Error())
		}
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
	log.Print(logs)
	if err != nil {
		return err
	}
	return SuccessResponse(c, fiber.StatusOK, "Logs retrieved successfully", fiber.Map{"logs": logs, "total": i})
}

func (h *LogsHandler) GetLogsMinMaxDates(c *fiber.Ctx) error {
	project := c.Params("project")
	if project == "" {
		return ErrorMessage(c, fiber.StatusBadRequest, "Project name is required")
	}
	exists, err := h.svc.CheckIfIndexExists(project)
	if err != nil {
		return InternalError(c, err)
	}
	if !exists {
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
func (h *LogsHandler) ListLogsFromArchive(c *fiber.Ctx) error {
	project := c.Params("project")
	if project == "" {
		return ErrorMessage(c, fiber.StatusBadRequest, "Project name is required")
	}
	exists, err := h.svc.CheckIfIndexExists(project)
	if err != nil {
		return InternalError(c, err)
	}
	if !exists {
		return ErrorMessage(c, fiber.StatusBadRequest, "project not found")
	}
	logs, err := h.svc.ListAllLogsFromStorage(project)
	if err != nil {
		return ErrorMessage(c, fiber.StatusInternalServerError, err.Error())
	}
	return SuccessResponse(c, fiber.StatusOK, "Logs retrieved successfully", fiber.Map{
		"logs": logs,
	})
}

func (h *LogsHandler) GetLogsFromColdStorage(c *fiber.Ctx) error {
	project := c.Params("project")
	fileName := c.Query("file")
	if fileName == "" {
		return ErrorMessage(c, fiber.StatusBadRequest, "File name is required")
	}
	level := c.Query("level")
	limit, err := strconv.Atoi(c.Query("limit", "100"))
	if err != nil {
		return ErrorMessage(c, fiber.StatusBadRequest, "invalid limit or page")
	}
	offset, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil {
		return ErrorMessage(c, fiber.StatusBadRequest, "invalid limit or page")
	}
	fromStr := c.Query("from", "")
	toStr := c.Query("to", "")
	sortByDate := c.Query("sortByDate", "ASC")
	sortByDate = strings.ToUpper(sortByDate)
	var sortEnum = []string{"ASC", "DESC"}
	if ok := slices.Contains(sortEnum, sortByDate); !ok {
		sortByDate = "ASC"
	}
	if project == "" {
		return ErrorMessage(c, fiber.StatusBadRequest, "project name is required")
	}
	var levelEnum = []string{"info", "debug", "warn", "error", "silly", "http", "verbose", ""}
	if ok := slices.Contains(levelEnum, level); !ok {
		return ErrorMessage(c, fiber.StatusBadRequest, "Invalid level")
	}
	fromFormatted, toFormatted, err := pkg.FormateFilterDateIfExists(fromStr, toStr)
	if err != nil {
		return ErrorMessage(c, fiber.StatusBadRequest, err.Error())
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

	exists, err := h.svc.CheckIfIndexExists(project)
	if err != nil {
		return InternalError(c, err)
	}
	if !exists {
		return ErrorMessage(c, fiber.StatusBadRequest, "project not found")
	}
	if archivedLogs, err := h.svc.ListAllLogsFromStorage(project); err != nil || !slices.Contains(archivedLogs, fileName) {
		return ErrorMessage(c, fiber.StatusNotFound, "File not found")
	}

	logs, totalCounts, err := h.svc.GetLogsFromArchive(project, fileName, filter)
	if err != nil {
		return ErrorMessage(c, fiber.StatusInternalServerError, err.Error())
	}
	return SuccessResponse(c, fiber.StatusOK, "Logs retrieved successfully", fiber.Map{
		"total": totalCounts,
		"logs":  logs,
	})
}

func (h *LogsHandler) StreamLogs(c *fiber.Ctx) error {
	project := c.Params("project")
	if project == "" {
		return ErrorMessage(c, fiber.StatusBadRequest, "Project name is required")
	}

	if _, err := h.svc.CheckIfIndexExists(project); err != nil {
		return ErrorMessage(c, fiber.StatusNotFound, "project not found")
	}

	user := c.Locals("user").(*pkg.UserClaims)
	clientID := user.UniqueName

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Transfer-Encoding", "chunked")

	// Register client
	h.sse.RegisterLogsClient(clientID, project)

	// Get a client channel
	clientChan, ok := h.sse.GetLogClientChannel(clientID)
	if !ok {
		log.Printf("No client channel found for ClientID: %s", clientID)
		h.sse.UnregisterLogsClient(clientID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get client channel"})
	}

	// Set up the streaming response
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer func() {
			log.Printf("Exiting stream writer for client: %s", clientID)
			h.sse.UnregisterLogsClient(clientID)
		}()

		// Send initial connection confirmation
		_, err := w.WriteString("data: connected\n\n")
		if err != nil {
			log.Printf("Error writing connection confirmation: %v", err)
			return
		}
		err = w.Flush()
		if err != nil {
			return
		}

		heartbeatTicker := time.NewTicker(30 * time.Second)
		defer heartbeatTicker.Stop()

		done := make(chan struct{})

		// Heartbeat goroutine
		go func() {
			for {
				select {
				case <-heartbeatTicker.C:
					// Send heartbeat
					_, err := w.WriteString("data: {\"type\":\"heartbeat\"}\n\n")
					if err != nil {
						log.Printf("Error writing heartbeat to client %s: %v", clientID, err)
						close(done)
						return
					}
					if err := w.Flush(); err != nil {
						log.Printf("Error flushing heartbeat for client %s: %v", clientID, err)
						close(done)
						return
					}
					// Update client activity
					h.sse.UpdateLogClientActivity(clientID)
				case <-done:
					return
				}
			}
		}()

		// Main message loop
		for {
			select {
			case logEntry, ok := <-clientChan:
				if !ok {
					log.Printf("Client channel closed for %s", clientID)
					return
				}

				data, err := json.Marshal(logEntry)
				if err != nil {
					log.Printf("Failed to marshal log entry: %v", err)
					continue
				}

				_, err = w.WriteString("data: " + string(data) + "\n\n")
				if err != nil {
					log.Printf("Error writing to client %s: %v", clientID, err)
					return
				}

				if err = w.Flush(); err != nil {
					log.Printf("Error flushing buffer for client %s: %v", clientID, err)
					return
				}

				// Update client activity
				h.sse.UpdateLogClientActivity(clientID)

			case <-done:
				return

			}
		}
	})

	return nil
}
