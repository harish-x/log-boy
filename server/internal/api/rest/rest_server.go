package rest

import (
	"context"
	"log"
	"server/config"
	"server/internal/api/rest/resthandlers"
	"server/internal/services"
	serversentevents "server/internal/services/server_sent_events"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func StartRestServer(ctx context.Context, cfg config.AppConfig, elasticSearch *elasticsearch.Client, ktm *config.KafkaTopicManager, LogSSE *services.SSEService, metricSSE *serversentevents.SSEMetricsService) error {
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

	synapse, err := config.NewSynapseSQL(cfg.SynapseDb, 10, 5, "1h")
	if err != nil {
		return err
	}

	restHandler := &resthandlers.RestHandler{
		App:           app,
		PostgresDb:    postgres,
		ElasticSearch: elasticSearch,
		SynapseDb:     synapse,
		Config:        cfg,
		Ktm:           ktm,
	}
	SetupRoutes(restHandler, LogSSE, metricSSE)
	go func() {
		<-ctx.Done()
		log.Println("Shutting down gRPC server...")
		err := ktm.Close()
		if err != nil {
			return
		}
		err = app.Shutdown()
		if err != nil {
			return
		}

	}()
	return app.Listen(cfg.ServerPort)
}

func SetupRoutes(h *resthandlers.RestHandler, l *services.SSEService, m *serversentevents.SSEMetricsService) {
	resthandlers.SetupProjectRoutes(h)
	resthandlers.SetupLogsRoutes(h, l)
	resthandlers.SetupMetricsHandler(h, m)
	resthandlers.SetupAlertRoutes(h)
}
