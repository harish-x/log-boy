package models

import (
	"time"

	"gorm.io/gorm"
)

type Alert struct {
	gorm.Model
	ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ProjectId  string    `json:"CheckIfProjectExistsByName" gorm:"type:uuid;not null"`
	Project    Project   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:ProjectId;references:ID"`
	RuleType   string    `json:"rule_type" gorm:"type:varchar(255)"`
	MetricName string    `json:"metric_name" gorm:"type:varchar(255)"`
	LogField   string    `json:"log_field" gorm:"type:varchar(255)"`
	Operator   string    `json:"operator" gorm:"type:varchar(255)"`
	Threshold  float32   `json:"threshold" gorm:"type:float"`
	TimeWindow string    `json:"time_window" gorm:"type:varchar(255)"`
	Status     string    `json:"status" gorm:"type:varchar(255); default:active"`
	Severity   string    `json:"severity" gorm:"type:varchar(255); default:info"`
	CreatedAt  time.Time `json:"created_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}

//CREATE TABLE alert_rules (
//id SERIAL PRIMARY KEY,
//project_id UUID REFERENCES projects(id),
//rule_type TEXT NOT NULL,              -- 'metric_avg', 'log_count', 'event_count'
//metric_name TEXT,                     -- e.g., 'cpu_usage', 'memory_usage'
//log_field TEXT,                       -- e.g., 'responseStatus'
//operator TEXT,                        -- e.g., '>', '<='
//threshold NUMERIC NOT NULL,           -- e.g., 80.0
//time_window INTERVAL NOT NULL,        -- e.g., '5 minutes'
//frequency INTERVAL DEFAULT '1 minute',
//filter_query JSONB DEFAULT NULL,      -- e.g., to filter by path/status/method
//severity TEXT CHECK (severity IN ('info', 'warning', 'critical')),
//status TEXT DEFAULT 'active',
//created_at TIMESTAMP DEFAULT now(),
//updated_at TIMESTAMP DEFAULT now()
//);
