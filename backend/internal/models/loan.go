package models

import (
	"time"

	"github.com/google/uuid"
)

// Loan status enum
const (
	LoanStatusPending  = "pending"
	LoanStatusApproved = "approved"
	LoanStatusRejected = "rejected"
	LoanStatusActive   = "active"   // funds disbursed
	LoanStatusClosed   = "closed"   // fully repaid
)

// Loan represents a loan application and lifecycle.
type Loan struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	NasabahID    uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"nasabah_id"`
	Amount       float64    `gorm:"not null;check:amount > 0"                      json:"amount"`
	InterestRate float64    `gorm:"not null;check:interest_rate > 0"               json:"interest_rate"`
	TermMonths   int        `gorm:"not null;check:term_months > 0"                 json:"term_months"`
	Status       string     `gorm:"not null;default:'pending'"                     json:"status"`
	ApprovedBy   *uuid.UUID `gorm:"type:uuid"                                      json:"-"` // hidden in safe DTO
	ApprovedAt   *time.Time `                                                      json:"approved_at,omitempty"`
	Notes        string     `gorm:"type:text"                                      json:"-"`
	CreatedAt    time.Time  `                                                      json:"created_at"`
	UpdatedAt    time.Time  `                                                      json:"updated_at"`

	// Relations
	Nasabah      *Nasabah      `gorm:"foreignKey:NasabahID"  json:"-"`
	Transactions []Transaction `gorm:"foreignKey:LoanID"     json:"-"`
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// LoanResponse — safe DTO; hides approver UUID and internal notes.
// OWASP API3 Secure.
type LoanResponse struct {
	ID           uuid.UUID  `json:"id"`
	NasabahID    uuid.UUID  `json:"nasabah_id"`
	Amount       float64    `json:"amount"`
	InterestRate float64    `json:"interest_rate"`
	TermMonths   int        `json:"term_months"`
	Status       string     `json:"status"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// LoanVulnerableResponse — INSECURE DTO; exposes approver UUID, notes, internals.
// TODO: Vulnerability Injection Point — OWASP API3 (BOPLA)
type LoanVulnerableResponse struct {
	ID           uuid.UUID  `json:"id"`
	NasabahID    uuid.UUID  `json:"nasabah_id"`
	Amount       float64    `json:"amount"`
	InterestRate float64    `json:"interest_rate"`
	TermMonths   int        `json:"term_months"`
	Status       string     `json:"status"`
	ApprovedBy   *uuid.UUID `json:"approved_by"`    // VULN: exposes internal staff UUID
	ApprovedAt   *time.Time `json:"approved_at"`
	Notes        string     `json:"notes"`          // VULN: internal processing notes
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ─── Request types ────────────────────────────────────────────────────────────

type CreateLoanRequest struct {
	Amount       float64 `json:"amount"        binding:"required,gt=0"`
	InterestRate float64 `json:"interest_rate" binding:"required,gt=0,lte=100"`
	TermMonths   int     `json:"term_months"   binding:"required,min=1,max=360"`
}

// UpdateLoanStatusRequest — used by staff/admin to update status.
// OWASP API3: In secure mode, 'approved' status is blocked for staff.
type UpdateLoanStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending approved rejected active closed"`
	Notes  string `json:"notes"`
}

func (l *Loan) ToResponse() LoanResponse {
	return LoanResponse{
		ID:           l.ID,
		NasabahID:    l.NasabahID,
		Amount:       l.Amount,
		InterestRate: l.InterestRate,
		TermMonths:   l.TermMonths,
		Status:       l.Status,
		ApprovedAt:   l.ApprovedAt,
		CreatedAt:    l.CreatedAt,
	}
}

func (l *Loan) ToVulnerableResponse() LoanVulnerableResponse {
	return LoanVulnerableResponse{
		ID:           l.ID,
		NasabahID:    l.NasabahID,
		Amount:       l.Amount,
		InterestRate: l.InterestRate,
		TermMonths:   l.TermMonths,
		Status:       l.Status,
		ApprovedBy:   l.ApprovedBy,
		ApprovedAt:   l.ApprovedAt,
		Notes:        l.Notes,
		CreatedAt:    l.CreatedAt,
		UpdatedAt:    l.UpdatedAt,
	}
}
