package resthandlers

import "github.com/gofiber/fiber/v2"

func ErrorMessage(ctx *fiber.Ctx, status int, message string) error {
	return ctx.Status(status).JSON(&fiber.Map{"message": message})
}

// InternalError handles internal server errors by setting the HTTP status to 500 and returning the error message as a JSON response.
func InternalError(ctx *fiber.Ctx, err error) error {
	return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
}

// BadRequestError sends a 400 Bad Request response with the given message as JSON.
func BadRequestError(ctx *fiber.Ctx, msg string) error {
	return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"message": msg,
	})
}

// SuccessResponse sends a JSON response with a specified status code, message, and data payload.
// It constructs the response using the provided Fiber context.
func SuccessResponse(ctx *fiber.Ctx, status int, message string, data interface{}) error {
	return ctx.Status(status).JSON(&fiber.Map{"message": message, "data": data})
}
