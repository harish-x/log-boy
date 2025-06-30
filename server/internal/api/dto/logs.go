package dto

import "time"

type Log struct {
	ServiceName  string       `json:"serviceName"`
	BuildDetails BuildDetails `json:"buildDetails"`
	Level        string       `json:"level"`
	Message      string       `json:"message"`
	Stack        string       `json:"stack"`
	Timestamp    time.Time    `json:"timestamp"`
}

type BuildDetails struct {
	NodeVersion string `json:"nodeVersion"`
	AppVersion  string `json:"appVersion"`
}

type LogFilter struct {
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
	Project       string `json:"project"`
	Level         string `json:"level"`
	From          string `json:"from"`
	To            string `json:"to"`
	SortByDate    string `json:"sortByDate"`
	LastTimestamp string `json:"lastTimestamp"`
}
