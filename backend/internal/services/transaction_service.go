package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/repository"
	"pt-dana-sejahtera/internal/security"
)

type TransactionService interface {
	Create(req models.CreateTransactionRequest, requestingUserID uuid.UUID, role string) (*models.TransactionResponse, error)
	ListByLoan(loanID uuid.UUID, requestingUserID uuid.UUID, role string, page, limit int) (interface{}, error)
	List(page, limit int) (interface{}, error)
}

type transactionService struct {
	txRepo      repository.TransactionRepository
	loanRepo    repository.LoanRepository
	nasabahRepo repository.NasabahRepository
	log         *zap.Logger
}

func NewTransactionService(
	txRepo repository.TransactionRepository,
	loanRepo repository.LoanRepository,
	nasabahRepo repository.NasabahRepository,
	log *zap.Logger,
) TransactionService {
	return &transactionService{txRepo: txRepo, loanRepo: loanRepo, nasabahRepo: nasabahRepo, log: log}
}

func (s *transactionService) Create(
	req models.CreateTransactionRequest,
	requestingUserID uuid.UUID,
	role string,
) (*models.TransactionResponse, error) {
	// A07 Secure: only admin/staff can record transactions.
	if security.IsSecureFor(security.CategoryA07) {
		if role == models.RoleNasabah {
			return nil, repository.ErrForbidden
		}
	}
	// TODO: Vulnerability Injection Point — OWASP API5 / A07 (Authentication Failures / BFLA)
	// A07 enabled: any authenticated user can create transactions.

	loanID, err := uuid.Parse(req.LoanID)
	if err != nil {
		return nil, errors.New("invalid loan_id")
	}

	// Verify loan exists
	loan, err := s.loanRepo.FindByID(loanID)
	if err != nil {
		return nil, fmt.Errorf("loan not found: %w", err)
	}

	// A06 Secure: enforce business-flow state machine on disbursement/repayment.
	if security.IsSecureFor(security.CategoryA06) {
		// Business rule: can only disburse approved loans
		if req.TransactionType == models.TxDisbursement && loan.Status != models.LoanStatusApproved {
			return nil, errors.New("can only disburse approved loans")
		}
		// Business rule: can only accept repayment on active loans
		if req.TransactionType == models.TxRepayment && loan.Status != models.LoanStatusActive {
			return nil, errors.New("loan is not in active status")
		}
	}
	// TODO: Vulnerability Injection Point — OWASP API6 / A06 (Vulnerable Components / Business Flow)
	// A06 enabled: no business flow validation — disburse or repay any loan in any state.

	tx := &models.Transaction{
		ID:              uuid.New(),
		LoanID:          loanID,
		Amount:          req.Amount,
		TransactionType: req.TransactionType,
		Description:     req.Description,
	}

	if err := s.txRepo.Create(tx); err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	s.log.Info("Transaction recorded",
		zap.String("tx_id", tx.ID.String()),
		zap.String("loan_id", loanID.String()),
		zap.String("type", req.TransactionType),
		zap.Float64("amount", req.Amount),
	)

	resp := tx.ToResponse()
	return &resp, nil
}

func (s *transactionService) ListByLoan(
	loanID uuid.UUID,
	requestingUserID uuid.UUID,
	role string,
	page, limit int,
) (interface{}, error) {
	if page < 1 {
		page = 1
	}

	loan, err := s.loanRepo.FindByID(loanID)
	if err != nil {
		return nil, err
	}

	// A01 Secure: verify nasabah owns this loan before listing its transactions.
	if security.IsSecureFor(security.CategoryA01) && role == models.RoleNasabah {
		nasabah, err := s.nasabahRepo.FindByUserID(requestingUserID)
		if err != nil || nasabah.ID != loan.NasabahID {
			s.log.Warn("BOLA attempt on transactions",
				zap.String("user_id", requestingUserID.String()),
				zap.String("loan_id", loanID.String()),
			)
			return nil, repository.ErrNotFound
		}
	}
	// TODO: Vulnerability Injection Point — OWASP API1 / A01 (BOLA)
	// A01 enabled: any authenticated user can list transactions of any loan.

	if limit < 1 || limit > maxSecureLimit {
		limit = defaultLimit
	}

	txs, total, err := s.txRepo.FindByLoanID(loanID, page, limit)
	if err != nil {
		return nil, err
	}

	result := make([]models.TransactionResponse, len(txs))
	for i, t := range txs {
		result[i] = t.ToResponse()
	}

	return map[string]interface{}{
		"data":  result,
		"total": total,
		"page":  page,
		"limit": limit,
	}, nil
}

func (s *transactionService) List(page, limit int) (interface{}, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > maxSecureLimit {
		limit = defaultLimit
	}

	txs, total, err := s.txRepo.List(page, limit)
	if err != nil {
		return nil, err
	}

	result := make([]models.TransactionResponse, len(txs))
	for i, t := range txs {
		result[i] = t.ToResponse()
	}

	return map[string]interface{}{
		"data":  result,
		"total": total,
		"page":  page,
		"limit": limit,
	}, nil
}
