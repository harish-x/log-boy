package resthandlers

import (
	"fmt"
	"server/internal/api/dto"
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
	api.Post("/", h.CreateAlert)
	api.Put("/:project/:id", h.UpdateAlert)
	api.Delete("/:project/:id", h.DeleteAlert)
	api.Put("/:project/email", h.CreateAlertEmail)
	api.Get("/:project/email", h.GetVerifiedEmail)
}

func (a *AlertHandler) GetAlerts(ctx *fiber.Ctx) error {
	return nil
}

func (a *AlertHandler) GetAlert(ctx *fiber.Ctx) error {
	return nil
}

func (a *AlertHandler) CreateAlert(ctx *fiber.Ctx) error {

	type RequestBody struct {
		RuleType     string  `json:"rule_type"`
		MetricName   string  `json:"metric_name"`
		Operator     string  `json:"operator"`
		Threshold    float32 `json:"threshold"`
		TimeWindow   string  `json:"time_window"`
		Severity     string  `json:"severity"`
		ProjectName  string  `json:"project_name"`
		AlertMethods []struct {
			Method string `json:"method"`
			Value  string `json:"value"`
		} `json:"alert_methods"`
	}

	var body RequestBody
	if err := ctx.BodyParser(&body); err != nil {
		return ErrorMessage(ctx, fiber.StatusBadRequest, "invalid alert payload: "+err.Error())
	}

	alert := models.Alert{
		ProjectName: body.ProjectName,
		RuleType:    body.RuleType,
		MetricName:  body.MetricName,
		Operator:    body.Operator,
		Threshold:   body.Threshold,
		TimeWindow:  body.TimeWindow,
		Severity:    body.Severity,
		Status:      "active",
	}

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
	type Email struct {
		Email string `json:"email"`
	}
	var Req Email

	if err := ctx.BodyParser(&Req); err != nil {
		return ErrorMessage(ctx, fiber.StatusBadRequest, "email is required")
	}
	verify, err := a.svc.VerifyEmail(Req.Email)
}

func (a *AlertHandler) GetVerifiedEmail(ctx *fiber.Ctx) error {
	project := ctx.Params("project")
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

	return SuccessResponse(ctx, fiber.StatusOK, "email retrived successfully", v)

}

func (a *AlertHandler) UpdateAlert(ctx *fiber.Ctx) error {
	return nil
}

func (a *AlertHandler) DeleteAlert(ctx *fiber.Ctx) error {
	return nil
}
