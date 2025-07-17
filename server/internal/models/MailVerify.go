package models

import (
	"time"

	"gorm.io/gorm"
)

type MailVerify struct {
	gorm.Model
	ID        uint      `json:"id" gorm:"primaryKey;type:int;unique"`
	Email     string    `json:"email" gorm:"type:varchar(255);not null"`
	OTP       string    `json:"otp" gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
