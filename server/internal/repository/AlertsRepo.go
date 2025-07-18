package repository

import (
	"server/internal/models"

	"github.com/elastic/go-elasticsearch/v9"
	"gorm.io/gorm"
)

type AlertsRepo interface {
	GetAlerts(project string) ([]*models.Alert, error)
	GetAlert(project, id string)
	CreateAlert(alert *models.Alert) (*models.Alert, error)
	UpdateAlert(alert *models.Alert) error
	DeleteAlert(id string) error
	CheckIfProjectExists(project string) (bool, error)
	GetVerifiedEmails(project string) ([]*models.VerifiedEmails, error)
	CreateEmailVerifyRequest(v *models.MailVerify) error
	VerifyEmail(email, project, otp string) (bool, error)
	CheckIsEmailVerified(email, project string) (bool, error)
	CreateAlertMethod(alertMethod *models.AlertMethods) error
}

type AlertRepo struct {
	es *elasticsearch.Client
	db *gorm.DB
}

func NewAlertRepo(es *elasticsearch.Client, db *gorm.DB) AlertsRepo {
	return &AlertRepo{es: es, db: db}
}
