package models

import (
	"time"

	"gorm.io/gorm"
)

type AlertMethods struct {
	gorm.Model
	ID          uint      `json:"id" gorm:"primaryKey;type:int;unique"`
	ProjectName string    `json:"project_name" gorm:"type:varchar(255);not null"`
	Project     Project   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:ProjectName;references:Name"`
	AlertID     string    `json:"alert_id" gorm:"type:uuid;not null"`
	Alert       Alert     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:AlertID;references:ID"`
	Method      string    `json:"method" gorm:"type:varchar(255)"`
	Value       string    `json:"value" gorm:"type:varchar(255)"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
