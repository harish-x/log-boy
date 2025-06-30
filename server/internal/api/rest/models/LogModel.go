package models

import (
	"time"
)

type BuildDetails struct {
	NodeVersion string `json:"nodeVersion"`
	AppVersion  string `json:"appVersion"`
}

type Log struct {
	ServiceName   string       `json:"serviceName" gorm:"type:nvarchar(255)"`
	BuildDetails  BuildDetails `json:"buildDetails" gorm:"type:nvarchar(max)"`
	Level         string       `json:"level" gorm:"type:nvarchar(50)"`
	Message       string       `json:"message" gorm:"type:text"`
	Stack         string       `json:"stack" gorm:"type:text"`
	RequestId     string       `json:"requestId" gorm:"type:nvarchar(255)"`
	RequestUrl    string       `json:"requestUrl" gorm:"type:text"`
	RequestMethod string       `json:"requestMethod" gorm:"type:nvarchar(10)"`
	IpAddress     string       `json:"ipAddress" gorm:"type:nvarchar(45)"`
	UserAgent     string       `json:"userAgent" gorm:"type:text"`
	Timestamp     time.Time    `json:"timestamp" gorm:"type:datetimeoffset"`
}
