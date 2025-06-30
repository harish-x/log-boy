package resthandlers

import "github.com/gofiber/fiber/v2"

func SetupProjectRoutes(r *RestHandler) {
	app := r.App

	app.Get("/projects", func(ctx *fiber.Ctx) error {
		return SuccessResponse(ctx, fiber.StatusOK, "success", nil)
	})
}
