package resthandlers

import (
	"server/internal/models"
	"server/internal/repository"
	"server/internal/services"

	"github.com/gofiber/fiber/v2"
)

type AlertHandler struct {
	svc *services.AlertServices
}

func SetupAlertRoutes(r *RestHandler) {
	app := r.App

	api := app.Group("/api/v1/alerts")
	svc := services.AlertServices{
		Repo: repository.NewAlertRepo(r.ElasticSearch, r.PostgresDb),
	}
	h := AlertHandler{
		svc: &svc,
	}

	api.Get("/:project/all", h.GetAlerts)
	api.Get("/:project/:id", h.GetAlert)
	api.Post("/:project", h.CreateAlert)
	api.Put("/:project/:id", h.UpdateAlert)
	api.Delete("/:project/:id", h.DeleteAlert)
	api.Put("/:project/email", h.CreateAlertEmail)
}

func (a *AlertHandler) GetAlerts(ctx *fiber.Ctx) error {
	return nil
}

func (a *AlertHandler) GetAlert(ctx *fiber.Ctx) error {
	return nil
}

func (a *AlertHandler) CreateAlert(ctx *fiber.Ctx) error {
	projectID := ctx.Params("project")
	var alert models.Alert
	var alertMethod []models.AlertMethods

	if err := ctx.BodyParser(&alert); err != nil {
		return ErrorMessage(ctx, fiber.StatusBadRequest, "invalid alert payload")
	}
	if err := ctx.BodyParser(&alertMethod); err != nil {
		return ErrorMessage(ctx, fiber.StatusBadRequest, "invalid alert method payload")
	}

	alert.ProjectId = projectID

	err := a.svc.CreateAlert(&alert, &alertMethod)

	if err != nil {
		return err
	}
	return SuccessResponse(ctx, fiber.StatusCreated, "alert created successfully", nil)
}

func (a *AlertHandler) CreateAlertEmail(ctx *fiber.Ctx) error {
	return nil
}

func (a *AlertHandler) UpdateAlert(ctx *fiber.Ctx) error {
	return nil
}

func (a *AlertHandler) DeleteAlert(ctx *fiber.Ctx) error {
	return nil
}
