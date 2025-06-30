package config

import (
	"context"
	"fmt"
	"log"
	"server/internal/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(addr string, maxOpenConns int, maxIdleConns int, maxConnLifetime string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(addr))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	pg, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get DB: %w", err)
	}
	pg.SetMaxOpenConns(maxOpenConns)
	pg.SetMaxIdleConns(maxIdleConns)
	duration, err := time.ParseDuration(maxConnLifetime)
	if err != nil {
		return nil, fmt.Errorf("invalid maxConnLifetime: %w", err)
	}
	pg.SetConnMaxLifetime(duration)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pg.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error; err != nil {
		return nil, fmt.Errorf("failed to create uuid extension: %w", err)
	}

	err = db.AutoMigrate(&models.Project{})
	if err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}
	log.Print("postgres connection established")
	return db, nil
}
