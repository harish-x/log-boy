package resthandlers

import (
	"server/config"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type RestHandler struct {
	App           *fiber.App
	PostgresDb    *gorm.DB
	Config        config.AppConfig
	ElasticSearch *elasticsearch.Client
	SynapseDb     *gorm.DB
	Ktm           *config.KafkaTopicManager
}
