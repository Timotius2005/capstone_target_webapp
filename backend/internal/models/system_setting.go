package models

import "time"

// SystemSettingsID is the fixed primary key for the singleton system_settings row.
// Application code always reads and writes this exact ID.
const SystemSettingsID = "00000000-0000-0000-0000-000000000001"

// SystemSetting stores the global runtime security mode that persists across restarts.
// The table contains exactly one row identified by SystemSettingsID.
type SystemSetting struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	Mode      string    `gorm:"type:varchar(20);not null;default:'secure'"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime"`
}

func (SystemSetting) TableName() string { return "system_settings" }

// ModeChangeLog records every mode-switch event for audit and forensic purposes.
type ModeChangeLog struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PreviousMode string    `gorm:"type:varchar(20);not null"`
	NewMode      string    `gorm:"type:varchar(20);not null"`
	IPAddress    string    `gorm:"type:varchar(45)"`
	UserAgent    string    `gorm:"type:text"`
	ChangedAt    time.Time `gorm:"not null;default:now()"`
}

func (ModeChangeLog) TableName() string { return "mode_change_logs" }
