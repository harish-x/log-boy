package services

import (
	"fmt"
	"log"
	"math/rand"
	"server/internal/api/dto"
	"server/internal/models"
	"server/internal/repository"
	"strconv"
	"time"
)

type AlertServices struct {
	Repo repository.AlertsRepo
}

func (as *AlertServices) CreateAlert(alert *models.Alert, alertMethods *[]models.AlertMethods) error {

	exists, err := as.Repo.CheckIfProjectExists(alert.ProjectName)
	if err != nil {
		return fmt.Errorf("project %s does not exist", alert.ProjectName)
	}
	if !exists {
		return fmt.Errorf("project %s does not exist", alert.ProjectName)
	}

	for _, alertMethod := range *alertMethods {
		if alertMethod.Method == "email" {
			verified, err := as.Repo.CheckIsEmailVerified(alertMethod.Value, alert.ProjectName)
			if err != nil {
				return fmt.Errorf("error while checking if email is verified: %w", err)
			}
			if !verified {
				return fmt.Errorf("email %s is not verified", alertMethod.Value)
			}
			if alertMethod.ProjectName != alert.ProjectName {
				return fmt.Errorf("email %s is not associated with project %s", alertMethod.Value, alert.ProjectName)
			}
			if alertMethod.ProjectName == "" {
				return fmt.Errorf("email %s is not associated with any project", alertMethod.Value)
			}
		}
	}
	newAlert, err := as.Repo.CreateAlert(alert)
	if err != nil {
		return err
	}
	for _, alertMethod := range *alertMethods {
		alertMethod.AlertID = newAlert.ID
		err := as.Repo.CreateAlertMethod(&alertMethod)
		if err != nil {
			return err
		}
	}
	return nil
}

func (as *AlertServices) GetVerifiedEmails(project string) ([]*models.VerifiedEmails, error) {
	exists, err := as.Repo.CheckIfProjectExists(project)
	if err != nil {
		return nil, fmt.Errorf("project %s does not exist", project)
	}
	if !exists {
		return nil, fmt.Errorf("project %s does not exist", project)
	}
	return as.Repo.GetVerifiedEmails(project)
}

func (as *AlertServices) RequestEmailVerify(p *dto.CreateVerifyEmail) error {
	ProjectExists, err := as.Repo.CheckIfProjectExists(p.Project)
	if err != nil || !ProjectExists {
		return fmt.Errorf("project %s does not exist", p.Project)
	}

	AlreadyVerified, err := as.Repo.CheckIsEmailVerified(p.Email, p.Project)
	if err != nil {
		return fmt.Errorf("internal Server Error")
	}
	if AlreadyVerified {
		return fmt.Errorf("email %s is already verified", p.Email)
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var OTP string
	OTP = ""
	for range 6 {
		OTP = OTP + strconv.Itoa(r.Intn(100))
	}

	verify := models.MailVerify{
		Email: p.Email,
		OTP:   OTP,
	}
	log.Print(OTP)
	err = as.Repo.CreateEmailVerifyRequest(&verify)
	if err != nil {
		return fmt.Errorf("internal Server Error")
	}
	return nil
}

func (as *AlertServices) VerifyEmail(email, project, otp string) (bool, error) {
	log.Print("im working", otp)
	return as.Repo.VerifyEmail(email, project, otp)
}
func (as *AlertServices) UpdateAlert(alert *models.Alert) error {
	err := as.Repo.UpdateAlert(alert)
	if err != nil {
		return err
	}
	return nil
}

func (as *AlertServices) GetAlerts(projectName string) ([]*models.Alert, error) {
	alerts, err := as.Repo.GetAlerts(projectName)
	if err != nil {
		return nil, err
	}
	return alerts, nil
}
func (as *AlertServices) DeleteAlert(id string) error {
	err := as.Repo.DeleteAlert(id)
	if err != nil {
		return err
	}
	return nil
}
