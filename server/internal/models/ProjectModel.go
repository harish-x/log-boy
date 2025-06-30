package models

import "time"

type Project struct {
	ID              string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name            string    `json:"name" gorm:"type:varchar(255)"`
	Description     string    `json:"description" gorm:"type:text"`
	Active          bool      `json:"active" gorm:"type:boolean;default:true"`
	ErrorRatio      float32   `json:"error_ratio" gorm:"type:real;default:0.0"`
	TotalErrors     int       `json:"total_errors" gorm:"type:integer;default:0"`
	TotalRequests   int       `json:"total_requests" gorm:"type:integer;default:0"`
	RetentionPeriod string    `json:"retention_period" gorm:"type:varchar(255)"`
	TotalWarnings   int       `json:"total_warnings" gorm:"type:integer;default:0"`
	CreatedAt       time.Time `json:"created_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
