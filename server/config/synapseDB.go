package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func NewSynapseSQL(addr string, maxOpenConns int, maxIdleConns int, maxConnLifetime string) (*gorm.DB, error) {

	// Open GORM DB with SQL Server driver
	db, err := gorm.Open(sqlserver.Open(addr))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	// Get a generic database object to configure a connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get generic DB object: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)

	duration, err := time.ParseDuration(maxConnLifetime)
	if err != nil {
		return nil, fmt.Errorf("invalid maxConnLifetime: %w", err)
	}
	sqlDB.SetConnMaxLifetime(duration)

	// Ping DB to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}
	log.Print("Synapse database is connected")
	return db, nil
}
