package models

type KeyStore struct {
	Key       string `json:"key" gorm:"primaryKey;type:varchar(255);not null"`
	Value     string `json:"value" gorm:"type:varchar(255);not null"`
	Timestamp int64  `json:"timestamp"`
}
