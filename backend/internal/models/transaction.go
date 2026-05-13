package models

import (
	"time"

	"github.com/google/uuid"
)

// Transaction type enum
const (
	TxDisbursement = "disbursement" // funds sent to nasabah
	TxRepayment    = "repayment"    // nasabah pays back
	TxPenalty      = "penalty"      // late payment fee
)

// Transaction records individual financial movements against a loan.
type Transaction struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LoanID          uuid.UUID `gorm:"type:uuid;not null;index"                       json:"loan_id"`
	Amount          float64   `gorm:"not null;check:amount > 0"                      json:"amount"`
	TransactionType string    `gorm:"not null"                                       json:"transaction_type"`
	Description     string    `gorm:"type:text"                                      json:"description"`
	CreatedAt       time.Time `                                                      json:"created_at"`

	// Relation
	Loan *Loan `gorm:"foreignKey:LoanID" json:"-"`
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type TransactionResponse struct {
	ID              uuid.UUID `json:"id"`
	LoanID          uuid.UUID `json:"loan_id"`
	Amount          float64   `json:"amount"`
	TransactionType string    `json:"transaction_type"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

// ─── Request types ────────────────────────────────────────────────────────────

type CreateTransactionRequest struct {
	LoanID          string  `json:"loan_id"          binding:"required,uuid"`
	Amount          float64 `json:"amount"           binding:"required,gt=0"`
	TransactionType string  `json:"transaction_type" binding:"required,oneof=disbursement repayment penalty"`
	Description     string  `json:"description"      binding:"max=500"`
}

func (t *Transaction) ToResponse() TransactionResponse {
	return TransactionResponse{
		ID:              t.ID,
		LoanID:          t.LoanID,
		Amount:          t.Amount,
		TransactionType: t.TransactionType,
		Description:     t.Description,
		CreatedAt:       t.CreatedAt,
	}
}
