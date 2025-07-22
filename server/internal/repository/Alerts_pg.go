package repository

import (
	"errors"
	"fmt"
	"log"
	"server/internal/models"
	"time"

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
	p := a.db.Model(&models.Project{}).Where("name = ?", project).First(&models.Project{})
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
	res := a.db.Model(&models.VerifiedEmails{}).Where("project_name = ?", project).Find(&emails)

	if res.Error != nil {
		return nil, errors.New("error while fetching emails")
	}

	return emails, nil
}

func (a *AlertRepo) CreateEmailVerifyRequest(v *models.MailVerify) error {
	var existing models.MailVerify

	// Check if email already exists
	err := a.db.Where("email = ?", v.Email).Order("created_at DESC").First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return a.db.Create(v).Error
	} else if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	existing.OTP = v.OTP
	existing.UpdatedAt = time.Now()

	return a.db.Save(&existing).Error
}

func (a *AlertRepo) VerifyEmail(email, project, otp string) (bool, error) {
	var verifyEmails models.MailVerify
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	err := a.db.
		Where("email = ? AND otp = ? AND updated_at >= ?", email, otp, fiveMinutesAgo).
		Order("created_at DESC").
		First(&verifyEmails).Error

	log.Printf("verifyEmails error: %v", err)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("db error: %w", err)
	}

	err = a.db.Create(&models.VerifiedEmails{
		ProjectName: project,
		Email:       email,
	}).Error

	if err != nil {
		return false, fmt.Errorf("db error: %w", err)
	}

	return true, nil
}
func (a *AlertRepo) CheckIsEmailVerified(email, project string) (bool, error) {
	var record models.VerifiedEmails
	err := a.db.Where("email = ? AND project_name = ?", email, project).First(&record).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("db error: %w", err)
	}

	return true, nil
}

func (a *AlertRepo) GetAlertRules(project string) ([]*models.Alert, error) {
	var alerts []*models.Alert
	err := a.db.Model(&models.Alert{}).Where("project_name = ?", project).Find(&alerts).Error
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}
	return alerts, err
}
