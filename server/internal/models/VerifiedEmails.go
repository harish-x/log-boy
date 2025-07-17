package models

import (
	"time"

	"gorm.io/gorm"
)

type VerifiedEmails struct {
	gorm.Model
	ID        uint      `json:"id" gorm:"primaryKey;type:int;unique"`
	ProjectID string    `json:"project_id" gorm:"type:uuid;not null"`
	Project   Project   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:ProjectID;references:ID"`
	Email     string    `json:"email" gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
