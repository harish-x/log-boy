package models

import (
	"time"
)

type BuildDetails struct {
	NodeVersion string `json:"nodeVersion"`
	AppVersion  string `json:"appVersion"`
}

type Log struct {
	ServiceName    string       `json:"serviceName" `
	BuildDetails   BuildDetails `json:"buildDetails"`
	Level          string       `json:"level"`
	Message        string       `json:"message"`
	Stack          string       `json:"stack"`
	RequestId      string       `json:"requestId"`
	RequestUrl     string       `json:"requestUrl"`
	RequestMethod  string       `json:"requestMethod"`
	IpAddress      string       `json:"ipAddress"`
	UserAgent      string       `json:"userAgent"`
	ResponseStatus string       `json:"responseStatus"`
	ResponseTime   string       `json:"responseTime"`
	Timestamp      time.Time    `json:"timestamp"`
}
