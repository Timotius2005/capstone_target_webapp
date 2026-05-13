package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role constants — fintech system has 3 tiers.
const (
	RoleAdmin   = "admin"   // full access, can approve loans
	RoleStaff   = "staff"   // can view/process, cannot approve
	RoleNasabah = "nasabah" // customer — own data only
)

// User is the auth entity.
type User struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Username      string         `gorm:"uniqueIndex;not null;size:100"                 json:"username"`
	Email         string         `gorm:"uniqueIndex;not null;size:255"                 json:"email"`
	PasswordHash  string         `gorm:"column:password_hash;not null"                 json:"-"`
	Role          string         `gorm:"not null;default:'nasabah'"                    json:"role"`
	IsActive      bool           `gorm:"not null;default:true"                         json:"-"`
	LoginAttempts int            `gorm:"not null;default:0"                            json:"-"`
	LastLoginAt   *time.Time     `                                                     json:"-"`
	CreatedAt     time.Time      `                                                     json:"created_at"`
	UpdatedAt     time.Time      `                                                     json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index"                                         json:"-"`

	// Relation
	Nasabah *Nasabah `gorm:"foreignKey:UserID" json:"-"`
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// UserResponse — safe public DTO; hides sensitive internal fields.
// OWASP API3 Secure: Broken Object Property Level Authorization.
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// UserVulnerableResponse — INSECURE DTO, exposes internals.
// TODO: Vulnerability Injection Point — OWASP API3 (BOPLA)
// Reveals password_hash, login_attempts, is_active flag.
type UserVulnerableResponse struct {
	ID            uuid.UUID  `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	PasswordHash  string     `json:"password_hash"`   // VULN: bcrypt hash exposed
	Role          string     `json:"role"`
	IsActive      bool       `json:"is_active"`
	LoginAttempts int        `json:"login_attempts"`  // VULN: internal counter
	LastLoginAt   *time.Time `json:"last_login_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

func (u *User) ToVulnerableResponse() UserVulnerableResponse {
	return UserVulnerableResponse{
		ID:            u.ID,
		Username:      u.Username,
		Email:         u.Email,
		PasswordHash:  u.PasswordHash,
		Role:          u.Role,
		IsActive:      u.IsActive,
		LoginAttempts: u.LoginAttempts,
		LastLoginAt:   u.LastLoginAt,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}
