package models

import (
	"time"
)

type Project struct {
	ID               string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name             string    `json:"name" gorm:"type:varchar(255); not null; unique"`
	Description      string    `json:"description" gorm:"type:text"`
	Active           bool      `json:"active" gorm:"type:boolean;default:true"`
	RetentionPeriod  string    `json:"retention_period" gorm:"type:varchar(255)"`
	ActiveMonitoring bool      `json:"active_monitoring" gorm:"type:boolean;default:false"`
	CreatedAt        time.Time `json:"created_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
