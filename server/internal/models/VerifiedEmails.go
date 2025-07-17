package models

import (
	"time"

	"gorm.io/gorm"
)

type VerifiedEmails struct {
	gorm.Model
	ID          uint      `json:"id" gorm:"primaryKey;type:int;unique"`
	ProjectName string    `json:"project_name" gorm:"type:varchar(255);not null"`
	Project     Project   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:ProjectName;references:Name"`
	Email       string    `json:"email" gorm:"type:varchar(255);not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
