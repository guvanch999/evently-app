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
				utc, _ := time.LoadLocation("")
				return time.Now().In(utc)
			},
			PrepareStmt:     false,
			CreateBatchSize: 100,
			Dialector:       postgres.New(postgres.Config{DSN: dsn}),
		},
	)
	if err != nil {
		return nil, err
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(40)
	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetConnMaxLifetime(time.Minute * 15)

	_, err = sqlDB.Conn(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error connection database: %v", err)
	}

	if !disableAutoMigration {
		err := database.AutoMigrate(
			&models.Users{},
		)
		if err != nil {
			return nil, fmt.Errorf("error migration database: %v", err)
		}
	}

	log.Println("Connected to PostgresSql")
	return database, nil
}
