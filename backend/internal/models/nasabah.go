package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Nasabah is the customer personal-data entity.
type Nasabah struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex"                 json:"user_id"`
	FullName    string         `gorm:"not null;size:255"                              json:"full_name"`
	NIK         string         `gorm:"not null;uniqueIndex;size:16"                   json:"-"` // masked in safe DTO
	Phone       string         `gorm:"size:20"                                        json:"phone"`
	Address     string         `gorm:"type:text"                                      json:"address"`
	DateOfBirth time.Time      `gorm:"not null"                                       json:"date_of_birth"`
	CreatedAt   time.Time      `                                                      json:"created_at"`
	UpdatedAt   time.Time      `                                                      json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                          json:"-"`

	// Relations
	User  *User  `gorm:"foreignKey:UserID" json:"-"`
	Loans []Loan `gorm:"foreignKey:NasabahID" json:"-"`
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// NasabahResponse — safe DTO; NIK is masked.
// OWASP API3 Secure: sensitive PII is hidden/masked.
type NasabahResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	FullName    string    `json:"full_name"`
	NIKMasked   string    `json:"nik"`     // e.g. "3201••••••••0001"
	Phone       string    `json:"phone"`
	Address     string    `json:"address"`
	DateOfBirth string    `json:"date_of_birth"`
	CreatedAt   time.Time `json:"created_at"`
}

// NasabahVulnerableResponse — INSECURE DTO; full NIK exposed.
// TODO: Vulnerability Injection Point — OWASP API3 (BOPLA)
// NIK (National ID) is sensitive PII — exposed in vulnerable mode.
type NasabahVulnerableResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	FullName    string    `json:"full_name"`
	NIK         string    `json:"nik"`           // VULN: full 16-digit NIK exposed
	Phone       string    `json:"phone"`
	Address     string    `json:"address"`
	DateOfBirth string    `json:"date_of_birth"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ─── Request types ────────────────────────────────────────────────────────────

type CreateNasabahRequest struct {
	FullName    string `json:"full_name"     binding:"required,min=2,max=255"`
	NIK         string `json:"nik"           binding:"required,len=16,numeric"`
	Phone       string `json:"phone"         binding:"required,min=8,max=20"`
	Address     string `json:"address"       binding:"required,min=5"`
	DateOfBirth string `json:"date_of_birth" binding:"required"` // YYYY-MM-DD
}

type UpdateNasabahRequest struct {
	FullName *string `json:"full_name" binding:"omitempty,min=2,max=255"`
	Phone    *string `json:"phone"     binding:"omitempty,min=8,max=20"`
	Address  *string `json:"address"   binding:"omitempty,min=5"`
}

// MaskNIK returns a partially hidden NIK: first 4 + •••••••• + last 4.
func MaskNIK(nik string) string {
	if len(nik) < 8 {
		return "••••••••••••••••"
	}
	return nik[:4] + "••••••••" + nik[len(nik)-4:]
}

func (n *Nasabah) ToResponse() NasabahResponse {
	return NasabahResponse{
		ID:          n.ID,
		UserID:      n.UserID,
		FullName:    n.FullName,
		NIKMasked:   MaskNIK(n.NIK),
		Phone:       n.Phone,
		Address:     n.Address,
		DateOfBirth: n.DateOfBirth.Format("2006-01-02"),
		CreatedAt:   n.CreatedAt,
	}
}

func (n *Nasabah) ToVulnerableResponse() NasabahVulnerableResponse {
	return NasabahVulnerableResponse{
		ID:          n.ID,
		UserID:      n.UserID,
		FullName:    n.FullName,
		NIK:         n.NIK, // VULN: full NIK
		Phone:       n.Phone,
		Address:     n.Address,
		DateOfBirth: n.DateOfBirth.Format("2006-01-02"),
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
}
