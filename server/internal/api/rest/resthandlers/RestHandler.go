package resthandlers

import (
	"server/config"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type RestHandler struct {
	App        *fiber.App
	PostgresDb *gorm.DB
	config     config.AppConfig
}
