package repository

import (
	"server/internal/api/dto"
	"server/internal/models"

	"github.com/elastic/go-elasticsearch/v9"
	"gorm.io/gorm"
)

type AlertsRepo interface {
	GetAlertRules(project string) ([]*models.Alert, error)
	GetAlerts(project string) (*[]dto.AlertMessage, error)
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
