package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/security"
)

// New opens a GORM/PostgreSQL connection.
func New(dsn string) (*gorm.DB, error) {
	logLevel := logger.Silent
	if security.IsVulnerable() {
		// TODO: Vulnerability Injection Point — OWASP API8 (Security Misconfiguration)
		// Debug mode logs raw SQL queries including sensitive data
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	return db, nil
}

// AutoMigrate runs GORM AutoMigrate for all domain models.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Nasabah{},
		&models.Loan{},
		&models.Transaction{},
	)
}
