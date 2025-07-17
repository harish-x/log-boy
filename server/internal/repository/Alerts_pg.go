package repository

import (
	"errors"
	"fmt"
	"server/internal/models"

	"gorm.io/gorm"
)

func (a *AlertRepo) CreateAlert(alert *models.Alert) (*models.Alert, error) {
	if err := a.db.Create(alert).Error; err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}
	return alert, nil
}

func (a *AlertRepo) CreateAlertMethod(alertMethod *models.AlertMethods) error {
	if err := a.db.Create(alertMethod).Error; err != nil {
		return fmt.Errorf("failed to create alert method: %w", err)
	}
	return nil
}

func (a *AlertRepo) UpdateAlert(alert *models.Alert) error {
	value := a.db.Save(alert)
	return value.Error
}

func (a *AlertRepo) DeleteAlert(id string) error {
	v := a.db.Delete(models.Alert{}, id)
	return v.Error
}

func (a *AlertRepo) CheckIfProjectExists(project string) (bool, error) {
	p := a.db.Model(&models.Project{}).Where("id = ?", project).First(&models.Project{})
	if p.Error == nil && errors.Is(p.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if p.Error != nil {
		return false, fmt.Errorf("db error: %w", p.Error)
	}
	return true, nil
}

func (a *AlertRepo) GetVerifiedEmails(project string) ([]*models.VerifiedEmails, error) {
	var emails []*models.VerifiedEmails
	res := a.db.Model(&models.VerifiedEmails{}).Where("project_id = ?", project).Find(&emails)

	if res.Error != nil {
		return nil, errors.New("error while fetching emails")
	}

	return emails, nil
}

func (a *AlertRepo) CheckIsEmailVerified(email string) (bool, error) {
	var record models.VerifiedEmails
	err := a.db.Where("email = ?", email).First(&record).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("db error: %w", err)
	}

	return true, nil
}
