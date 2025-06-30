package models

import "time"

type Project struct {
	ID              string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Active          bool      `json:"active"  gorm:"column:active"`
	ErrorRatio      float32   `json:"error_ratio"`
	TotalErrors     int       `json:"total_errors"`
	TotalRequests   int       `json:"total_requests"`
	RetentionPeriod string    `json:"retention_period"`
	TotalWarnings   int       `json:"total_warnings"`
	CreatedAt       time.Time `json:"created_at"  gorm:"<-:false"`
	UpdatedAt       time.Time `json:"updated_at"`
}
