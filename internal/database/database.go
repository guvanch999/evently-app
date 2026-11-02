package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/guvanch999/evently-app/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func Connect(dsn string, disableAutoMigration bool) (*gorm.DB, error) {
	database, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogger.Warn),
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
			PrepareStmt:     false,
			CreateBatchSize: 100,
		},
	)
	if err != nil {
		return nil, err
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Minute * 15)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if !disableAutoMigration {
		if err := database.AutoMigrate(models.All()...); err != nil {
			return nil, fmt.Errorf("error migrating database: %w", err)
		}
	}

	log.Println("Connected to PostgreSQL")
	return database, nil
}
