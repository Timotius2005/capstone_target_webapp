package usecase

import (
	"time"

	"github.com/google/uuid"
	"pt-dana-sejahtera/internal/domain"
	"pt-dana-sejahtera/internal/repository"
)

type LoanUseCase struct {
	loanRepo *repository.LoanRepository
}

func NewLoanUseCase(loanRepo *repository.LoanRepository) *LoanUseCase {
	return &LoanUseCase{loanRepo: loanRepo}
}

type CreateLoanRequest struct {
	NasabahID    uuid.UUID `json:"nasabah_id"`
	Amount       float64   `json:"amount"`
	InterestRate float64   `json:"interest_rate"`
	TermMonths   int       `json:"term_months"`
}

func (uc *LoanUseCase) GetAllLoans() ([]*domain.Loan, error) {
	// TODO: SECURITY VULNERABILITY - Excessive Data Exposure
	// Should not return all loans without pagination and filtering
	return uc.loanRepo.GetAll()
}

func (uc *LoanUseCase) GetLoanByID(id uuid.UUID) (*domain.Loan, error) {
	// TODO: SECURITY VULNERABILITY - Broken Object Level Authorization
	// Should check if user has access to this loan
	return uc.loanRepo.GetByID(id)
}

func (uc *LoanUseCase) CreateLoan(req CreateLoanRequest) error {
	loan := &domain.Loan{
		ID:           uuid.New(),
		NasabahID:    req.NasabahID,
		Amount:       req.Amount,
		InterestRate: req.InterestRate,
		TermMonths:   req.TermMonths,
		Status:       "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return uc.loanRepo.Create(loan)
}