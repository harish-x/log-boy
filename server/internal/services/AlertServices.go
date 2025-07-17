package services

import (
	"fmt"
	"server/internal/models"
	"server/internal/repository"
)

type AlertServices struct {
	Repo repository.AlertsRepo
}

func (as *AlertServices) CreateAlert(alert *models.Alert, alertMethods *[]models.AlertMethods) error {

	exists, err := as.Repo.CheckIfProjectExists(alert.ProjectId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("project %s does not exist", alert.ProjectId)
	}

	for _, alertMethod := range *alertMethods {
		if alertMethod.Method == "email" {
			verified, err := as.Repo.CheckIsEmailVerified(alertMethod.Value)
			if err != nil {
				return fmt.Errorf("error while checking if email is verified: %w", err)
			}
			if !verified {
				return fmt.Errorf("email %s is not verified", alertMethod.Value)
			}
			if alertMethod.ProjectID != alert.ProjectId {
				return fmt.Errorf("email %s is not associated with project %s", alertMethod.Value, alert.ProjectId)
			}
			if alertMethod.ProjectID == "" {
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

func (as *AlertServices) UpdateAlert(alert *models.Alert) error {
	err := as.Repo.UpdateAlert(alert)
	if err != nil {
		return err
	}
	return nil
}

func (as *AlertServices) DeleteAlert(id string) error {
	err := as.Repo.DeleteAlert(id)
	if err != nil {
		return err
	}
	return nil
}
