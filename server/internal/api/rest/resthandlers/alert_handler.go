package resthandlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"server/internal/api/dto"
	"server/internal/models"
	"server/internal/repository"
	"server/internal/services"
	serversentevents "server/internal/services/server_sent_events"
	"server/pkg"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AlertHandler struct {
	svc *services.AlertServices
	sse *serversentevents.SSEAlertService
}

func SetupAlertRoutes(r *RestHandler, a *serversentevents.SSEAlertService) {
	app := r.App

	api := app.Group("/api/v1/alerts")
	svc := services.AlertServices{
		Repo: repository.NewAlertRepo(r.ElasticSearch, r.PostgresDb),
	}
	h := AlertHandler{
		svc: &svc,
		sse: a,
	}

	api.Post("/email", pkg.AuthMiddleware(), h.CreateAlertEmail)
	api.Patch("/email/verify", pkg.AuthMiddleware(), h.VerifyEmail)
	api.Post("/new", pkg.AuthMiddleware(), h.CreateAlert)
	api.Get("/email/:project", pkg.AuthMiddleware(), h.GetVerifiedEmail)
	api.Get("/:project/all", pkg.AuthMiddleware(), h.GetAlertRules)
	api.Get("/:project/stream", pkg.SSEAuthMiddleware(), h.SendAlert)
	//api.Get("/:project/:id", h.GetAlert)
	//api.Put("/:project/:id", h.UpdateAlert)
	//api.Delete("/:project/:id", h.DeleteAlert)
	api.Get("/:project/old_alerts", pkg.AuthMiddleware(), h.GetAlerts)
}

func (a *AlertHandler) GetAlertRules(ctx *fiber.Ctx) error {
	project := ctx.Params("project")
	if project == "" {
		return ErrorMessage(ctx, fiber.StatusBadRequest, "project name is required")
	}
	alerts, err := a.svc.GetAlertRules(project)
	if err != nil {
		return ErrorMessage(ctx, fiber.StatusInternalServerError, "Internal server error")
	}

	return SuccessResponse(ctx, fiber.StatusOK, "alerts retrieved successfully", alerts)
}

func (a *AlertHandler) GetAlerts(ctx *fiber.Ctx) error {
	project := ctx.Params("project")
	if project == "" {
		return ErrorMessage(ctx, fiber.StatusBadRequest, "project name is required")
	}
	alerts, err := a.svc.GetAlerts(project)

	if err != nil {
		return ErrorMessage(ctx, fiber.StatusInternalServerError, "Internal server error")
	}
	return SuccessResponse(ctx, fiber.StatusOK, "alerts retrieved successfully", alerts)

}

func (a *AlertHandler) CreateAlert(ctx *fiber.Ctx) error {

	type RequestBody struct {
		RuleType      string  `json:"rule_type"`
		MetricName    string  `json:"metric_name"`
		Operator      string  `json:"operator"`
		Threshold     float32 `json:"threshold"`
		TimeWindow    string  `json:"time_window"`
		Severity      string  `json:"severity"`
		ProjectName   string  `json:"project_name"`
		LogField      string  `json:"log_field"`
		LogFieldValue string  `json:"log_field_value"`
		AlertMethods  []struct {
			Method string `json:"method"`
			Value  string `json:"value"`
		} `json:"alert_methods"`
	}

	var body RequestBody
	if err := ctx.BodyParser(&body); err != nil {
		return ErrorMessage(ctx, fiber.StatusBadRequest, "invalid alert payload: "+err.Error())
	}

	alert := models.Alert{
		ProjectName:   body.ProjectName,
		RuleType:      body.RuleType,
		MetricName:    body.MetricName,
		Operator:      body.Operator,
		Threshold:     body.Threshold,
		TimeWindow:    body.TimeWindow,
		LogField:      body.LogField,
		LogFieldValue: body.LogFieldValue,
		Severity:      body.Severity,
		Status:        "active",
	}
	log.Print(alert)
	var alertMethods []models.AlertMethods
	for _, method := range body.AlertMethods {
		alertMethods = append(alertMethods, models.AlertMethods{
			ProjectName: body.ProjectName,
			Method:      method.Method,
			Value:       method.Value,
		})
	}
	err := a.svc.CreateAlert(&alert, &alertMethods)

	if err != nil {
		return err
	}
	return SuccessResponse(ctx, fiber.StatusCreated, "alert created successfully", nil)
}

func (a *AlertHandler) CreateAlertEmail(ctx *fiber.Ctx) error {

	var reqBody dto.CreateVerifyEmail

	if err := ctx.BodyParser(&reqBody); err != nil {
		return ErrorMessage(ctx, fiber.StatusBadRequest, "invalid payload")
	}
	err := a.svc.RequestEmailVerify(&reqBody)

	if err != nil && err.Error() == "Internal Server Error" {
		return ErrorMessage(ctx, fiber.StatusInternalServerError, err.Error())
	}
	if err != nil {
		return ErrorMessage(ctx, fiber.StatusBadRequest, err.Error())
	}
	successMsg := fmt.Sprintf("otp sent to %s", reqBody.Email)
	return SuccessResponse(ctx, fiber.StatusCreated, successMsg, nil)
}

func (a *AlertHandler) VerifyEmail(ctx *fiber.Ctx) error {

	type Verify struct {
		Email   string `json:"email"`
		Project string `json:"project"`
		OTP     string `json:"otp"`
	}
	var Req Verify

	if err := ctx.BodyParser(&Req); err != nil {
		return ErrorMessage(ctx, fiber.StatusBadRequest, "email is required")
	}
	log.Print(Req.Email, Req.Project, Req.OTP)
	verify, err := a.svc.VerifyEmail(Req.Email, Req.Project, Req.OTP)
	if err != nil {
		return ErrorMessage(ctx, fiber.StatusInternalServerError, "Internal server error")
	}
	if verify {
		return SuccessResponse(ctx, fiber.StatusOK, "email verified successfully", nil)
	}
	return ErrorMessage(ctx, fiber.StatusBadRequest, "email not verified")
}

func (a *AlertHandler) GetVerifiedEmail(ctx *fiber.Ctx) error {
	project := ctx.Params("project")
	log.Print("im working")
	if project == "" {
		return ErrorMessage(ctx, fiber.StatusBadRequest, "project name is required")
	}
	type verifiedMails struct {
		Email string `json:"email"`
	}
	var v []verifiedMails
	mails, err := a.svc.GetVerifiedEmails(project)
	if err != nil {
		return ErrorMessage(ctx, fiber.StatusInternalServerError, "Internal server error")
	}

	for _, m := range mails {
		var c verifiedMails
		c.Email = m.Email
		v = append(v, c)
	}

	return SuccessResponse(ctx, fiber.StatusOK, "email retrieved successfully", v)

}

func (a *AlertHandler) SendAlert(c *fiber.Ctx) error {
	project := c.Params("project")
	if project == "" {
		return ErrorMessage(c, fiber.StatusBadRequest, "Project name is required")
	}

	projectExists, err := a.svc.CheckProjectExists(project)
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
	a.sse.RegisterAlertClient(clientID, project)

	// Get a client channel
	clientChan, ok := a.sse.GetAlertClientChannel(clientID)
	if !ok {
		log.Printf("No client channel found for ClientID: %s", clientID)
		a.sse.UnregisterAlertClient(clientID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get client channel"})
	}

	// Set up the streaming response
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer func() {
			log.Printf("Exiting stream writer for client: %s", clientID)
			a.sse.UnregisterAlertClient(clientID)
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
					a.sse.UpdateAlertClientActivity(clientID)
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
				a.sse.UpdateAlertClientActivity(clientID)

			case <-done:
				return

			}
		}
	})

	return nil
}

func (a *AlertHandler) UpdateAlert(ctx *fiber.Ctx) error {
	return nil
}

func (a *AlertHandler) DeleteAlert(ctx *fiber.Ctx) error {
	return nil
}
