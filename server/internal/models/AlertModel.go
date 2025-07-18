package models

import (
	"time"
)

type Alert struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ProjectName string    `json:"project_name" gorm:"type:varchar(255);not null"`
	Project     Project   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:ProjectName;references:Name" json:"-"`
	RuleType    string    `json:"rule_type" gorm:"type:varchar(255)"`
	MetricName  string    `json:"metric_name" gorm:"type:varchar(255)"`
	LogField    string    `json:"log_field" gorm:"type:varchar(255)"`
	Operator    string    `json:"operator" gorm:"type:varchar(255)"`
	Threshold   float32   `json:"threshold" gorm:"type:float"`
	TimeWindow  string    `json:"time_window" gorm:"type:varchar(255)"`
	Status      string    `json:"status" gorm:"type:varchar(255); default:active"`
	Severity    string    `json:"severity" gorm:"type:varchar(255); default:info"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
