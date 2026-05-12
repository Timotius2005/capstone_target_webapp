package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Nasabah struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	NIK         string    `json:"nik" db:"nik"`
	Phone       string    `json:"phone" db:"phone"`
	Address     string    `json:"address" db:"address"`
	DateOfBirth time.Time `json:"date_of_birth" db:"date_of_birth"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Loan struct {
	ID            uuid.UUID `json:"id" db:"id"`
	NasabahID     uuid.UUID `json:"nasabah_id" db:"nasabah_id"`
	Amount        float64   `json:"amount" db:"amount"`
	InterestRate  float64   `json:"interest_rate" db:"interest_rate"`
	TermMonths    int       `json:"term_months" db:"term_months"`
	Status        string    `json:"status" db:"status"`
	ApprovedAt    *time.Time `json:"approved_at" db:"approved_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type Transaction struct {
	ID          uuid.UUID `json:"id" db:"id"`
	LoanID      uuid.UUID `json:"loan_id" db:"loan_id"`
	Type        string    `json:"type" db:"type"`
	Amount      float64   `json:"amount" db:"amount"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}