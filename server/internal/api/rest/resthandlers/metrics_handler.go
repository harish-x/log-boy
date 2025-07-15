package resthandlers

import (
	"bufio"
	"encoding/json"
	"log"
	"server/internal/repository"
	"server/internal/services"
	serversentevents "server/internal/services/server_sent_events"
	"server/pkg"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type MetricsHandler struct {
	sse *serversentevents.SSEMetricsService
	svc *services.MetricsServices
}

func SetupMetricsHandler(r *RestHandler, l *serversentevents.SSEMetricsService) {
	app := r.App
	svc := services.MetricsServices{
		Repo:       repository.NewMetricsRepo(r.ElasticSearch),
		ProjetRepo: repository.NewProjectRepo(r.PostgresDb),
	}
	handler := MetricsHandler{
		sse: l,
		svc: &svc,
	}
	api := app.Group("/api/v1/metrics/")

	api.Get("/:project/stream", pkg.SSEAuthMiddleware(), handler.StreamMetrics)
	api.Get("/:project/cpu", handler.GetCpuUsage)
	api.Get("/:project/memory", handler.Getmemoryusage)
	api.Get("/:project/date", handler.GetMetricsMinMaxDates)

}

func (h *MetricsHandler) StreamMetrics(c *fiber.Ctx) error {
	project := c.Params("project")
	if project == "" {
		return ErrorMessage(c, fiber.StatusBadRequest, "Project name is required")
	}

	projectExists, err := h.svc.CheckIfProjectExists(project)
	if err != nil {
		return ErrorMessage(c, fiber.StatusInternalServerError, err.Error())
	}
	if !projectExists {
		return ErrorMessage(c, fiber.StatusNotFound, "Project not found")
	}
	user := c.Locals("user").(*pkg.UserClaims)
	clientID := user.UniqueName

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Transfer-Encoding", "chunked")

	// Register client
	h.sse.RegisterMetricsClient(clientID, project)

	// Get a client channel
	clientChan, ok := h.sse.GetMetricsClientChannel(clientID)
	if !ok {
		log.Printf("No client channel found for ClientID: %s", clientID)
		h.sse.UnRegisterMetricsClient(clientID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get client channel"})
	}

	// Set up the streaming response
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer func() {
			log.Printf("Exiting stream writer for client: %s", clientID)
			h.sse.UnRegisterMetricsClient(clientID)
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
					h.sse.UpdateMetricsClientActivity(clientID)
				case <-done:
					return
				}
			}
		}()

		// Main message loop
		for {
			select {
			case metrics, ok := <-clientChan:
				if !ok {
					log.Printf("Client channel closed for %s", clientID)
					return
				}

				data, err := json.Marshal(metrics)
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
				h.sse.UpdateMetricsClientActivity(clientID)

			case <-done:
				return

			}
		}
	})

	return nil
}

func (h *MetricsHandler) GetCpuUsage(c *fiber.Ctx) error {
	project := c.Params("project")
	if project == "" {
		return ErrorMessage(c, fiber.StatusBadRequest, "Project name is required")
	}
	from := c.Query("from")
	to := c.Query("to", "0")
	groupBy := c.Query("points")
	var formattedFrom int64
	var formettedTo int64 = 0
	var err error

	projectExists, err := h.svc.CheckIfProjectExists(project)
	if err != nil {
		return ErrorMessage(c, fiber.StatusInternalServerError, err.Error())
	}
	if !projectExists {
		return ErrorMessage(c, fiber.StatusNotFound, "Project not found")
	}

	if from != "" {
		formattedFrom, err = strconv.ParseInt(from, 10, 64)
		if err != nil {
			return ErrorMessage(c, fiber.StatusBadRequest, "invalid from date")
		}
	}

	if to != "0" {
		formettedTo, err = strconv.ParseInt(to, 10, 64)
		if err != nil {
			return ErrorMessage(c, fiber.StatusBadRequest, "invalid to date")
		}
	}
	log.Printf("Formatted From: %d, Formatted To: %d", formattedFrom, formettedTo)

	res, err := h.svc.GetCpuUsage(project, formattedFrom, formettedTo, groupBy)

	if err != nil {
		return ErrorMessage(c, fiber.StatusInternalServerError, err.Error())
	}

	return SuccessResponse(c, fiber.StatusOK, "success", res)
}

func (h *MetricsHandler) Getmemoryusage(c *fiber.Ctx) error {
	project := c.Params("project")
	if project == "" {
		return ErrorMessage(c, fiber.StatusBadRequest, "Project name is required")
	}
	from := c.Query("from")
	to := c.Query("to")
	groupBy := c.Query("points")

	projectExists, err := h.svc.CheckIfProjectExists(project)
	if err != nil {
		return ErrorMessage(c, fiber.StatusInternalServerError, err.Error())
	}
	if !projectExists {
		return ErrorMessage(c, fiber.StatusNotFound, "Project not found")
	}

	formattedFrom, err := strconv.ParseInt(from, 10, 64)
	if err != nil {
		return ErrorMessage(c, fiber.StatusBadRequest, "invalid from date")
	}
	formettedTo, err := strconv.ParseInt(to, 10, 64)
	if err != nil {
		return ErrorMessage(c, fiber.StatusBadRequest, "invalid to date")
	}
	log.Printf("Formatted From: %d, Formatted To: %d", formattedFrom, formettedTo)
	res, err := h.svc.GetMemoryUsage(project, formattedFrom, formettedTo, groupBy)

	if err != nil {
		return ErrorMessage(c, fiber.StatusInternalServerError, err.Error())
	}

	return SuccessResponse(c, fiber.StatusOK, "success", res)
}

func (h *MetricsHandler) GetMetricsMinMaxDates(c *fiber.Ctx) error {
	project := c.Params("project")
	if project == "" {
		return ErrorMessage(c, fiber.StatusBadRequest, "Project name is required")
	}
	exists, err := h.svc.CheckIfProjectExists(project)
	if err != nil {
		return InternalError(c, err)
	}
	if !exists {
		return ErrorMessage(c, fiber.StatusNotFound, "Project not found")
	}
	dates, err := h.svc.GetMetricsMinMaxDate(project)
	if err != nil {
		return ErrorMessage(c, fiber.StatusNotFound, err.Error())
	}
	return SuccessResponse(c, fiber.StatusOK, "Metrics dates retrieved successfully", dates)
}
