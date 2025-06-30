package rest

import (
	"context"
	"server/config"
	"server/internal/api/rest/resthandlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func StartRestServer(ctx context.Context, cfg config.AppConfig) error {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:  "http://localhost:5173",
		AllowHeaders:  "Origin,Content-Type,Accept,Authorization",
		AllowMethods:  "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		ExposeHeaders: "Content-Length, Content-Type, Content-Disposition, X-Content-Type-Options, X-Frame-Options, X-XSS-Protection",
	}))
	postgres, err := config.NewPostgres(cfg.PostgresDb, 10, 5, "1h")
	if err != nil {
		return err
	}
	elasticSearch, err := config.NewElasticSearchDB(cfg.ElasticSearch)
	if err != nil {
		return err
	}
	restHandler := &resthandlers.RestHandler{
		App:           app,
		PostgresDb:    postgres,
		ElasticSearch: elasticSearch,
	}
	SetupRoutes(restHandler)
	return app.Listen(cfg.ServerPort)
}

func SetupRoutes(h *resthandlers.RestHandler) {
	resthandlers.SetupProjectRoutes(h)
	resthandlers.SetupLogsRoutes(h)
}
