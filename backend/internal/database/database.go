package database

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/security"
)

const (
	maxRetries  = 10
	retryDelay  = 3 * time.Second
	maxOpenConn = 25
	maxIdleConn = 10
)

// New opens a GORM/PostgreSQL connection with retry logic.
// It is the single point of database initialization — do not create additional
// connection functions elsewhere.
//
// Retries up to maxRetries times with retryDelay between attempts.
// Returns a ready-to-use *gorm.DB or an error with debugging hints.
func New(dsn string, log *zap.Logger) (*gorm.DB, error) {
	gormLogLevel := logger.Silent
	if security.IsVulnerable() {
		// TODO: Vulnerability Injection Point — OWASP API8 (Security Misconfiguration)
		// Debug mode logs raw SQL including sensitive query parameters
		gormLogLevel = logger.Info
	}

	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	}

	var (
		db      *gorm.DB
		lastErr error
	)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, lastErr = gorm.Open(postgres.Open(dsn), gormCfg)
		if lastErr == nil {
			// Verify the connection is actually alive
			sqlDB, pingErr := db.DB()
			if pingErr == nil {
				pingErr = sqlDB.Ping()
			}
			if pingErr == nil {
				// Connection is good
				sqlDB.SetMaxOpenConns(maxOpenConn)
				sqlDB.SetMaxIdleConns(maxIdleConn)
				sqlDB.SetConnMaxLifetime(30 * time.Minute)
				return db, nil
			}
			lastErr = pingErr
		}

		log.Warn("Database connection attempt failed",
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", maxRetries),
			zap.Duration("next_retry_in", retryDelay),
			zap.Error(lastErr),
		)

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	// All attempts exhausted — print actionable debug hints
	log.Error("Database connection failed after all attempts — debugging checklist:",
		zap.Int("attempts_made", maxRetries),
		zap.Error(lastErr),
	)
	log.Error("  1. Is the Docker container running?    → docker ps | grep fintech-postgres")
	log.Error("  2. Is port 5432 exposed?               → docker port fintech-postgres 5432")
	log.Error("  3. Check DB_HOST value                 → 'localhost' (local) or VM IP (remote)")
	log.Error("  4. Check firewall / UFW                → sudo ufw status")
	log.Error("  5. Test connectivity manually          → psql -h $DB_HOST -U fintech -d nasabahdb")

	return nil, fmt.Errorf("database: failed after %d attempts: %w", maxRetries, lastErr)
}

// AutoMigrate creates or updates all domain tables.
// Must be called once after New() succeeds.
func AutoMigrate(db *gorm.DB, log *zap.Logger) error {
	log.Info("Running database migrations...")

	if err := db.AutoMigrate(
		&models.User{},
		&models.Nasabah{},
		&models.Loan{},
		&models.Transaction{},
		&models.SystemSetting{},
		&models.ModeChangeLog{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	log.Info("Database migration successful",
		zap.Strings("tables", []string{
			"users", "nasabah", "loans", "transactions",
			"system_settings", "mode_change_logs",
		}),
	)
	return nil
}

// LoadOrInitModeFromDB loads the persisted security mode from system_settings.
// If no row exists yet, it seeds one using fallbackMode (from APP_SECURITY_MODE env).
// The in-memory security mode is updated to match the database value.
// Call this once after AutoMigrate succeeds.
func LoadOrInitModeFromDB(db *gorm.DB, log *zap.Logger, fallbackMode string) {
	var setting models.SystemSetting
	result := db.Where("id = ?", models.SystemSettingsID).First(&setting)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// No row yet — seed from the env-var default.
		dbMode := "secure"
		if fallbackMode == "sandbox" || fallbackMode == "vulnerable" {
			dbMode = "vulnerable"
		}
		setting = models.SystemSetting{
			ID:        models.SystemSettingsID,
			Mode:      dbMode,
			UpdatedAt: time.Now().UTC(),
		}
		if err := db.Create(&setting).Error; err != nil {
			log.Warn("Failed to seed system_settings row — using in-memory default",
				zap.Error(err),
			)
			return
		}
		log.Info("system_settings row seeded", zap.String("mode", dbMode))
	} else if result.Error != nil {
		log.Warn("Failed to load mode from DB — keeping env-var default",
			zap.Error(result.Error),
		)
		return
	}

	// Apply DB value to the in-memory security package (overrides env-var Init).
	if setting.Mode == "vulnerable" {
		security.SetMode(security.ModeSandbox)
	} else {
		security.SetMode(security.ModeSecure)
	}
	log.Info("Security mode loaded from database",
		zap.String("mode", setting.Mode),
		zap.String("note", "DB value overrides APP_SECURITY_MODE env var"),
	)
}

// HealthCheck returns nil if the database is reachable, or an error otherwise.
// Used by the /health/db endpoint.
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}
