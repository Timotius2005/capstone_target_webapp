package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/repository"
	"pt-dana-sejahtera/internal/security"
)

// Business rules
const (
	MaxPendingLoansPerNasabah = 3  // OWASP API6: max concurrent pending applications
	LoanCreationWindowSec     = 60 // time window for rate limiting
	MaxLoansPerWindow         = 2  // OWASP API4: max loans per window
)

type LoanService interface {
	Apply(req models.CreateLoanRequest, userID uuid.UUID) (interface{}, error)
	GetByID(loanID uuid.UUID, requestingUserID uuid.UUID, role string) (interface{}, error)
	List(userID uuid.UUID, role string, page, limit int) (interface{}, error)
	Approve(loanID uuid.UUID, approverUserID uuid.UUID, role string) (interface{}, error)
	Reject(loanID uuid.UUID, staffUserID uuid.UUID, role string, notes string) (interface{}, error)
	UpdateStatus(loanID uuid.UUID, req models.UpdateLoanStatusRequest, requestingUserID uuid.UUID, role string) (interface{}, error)
}

type loanService struct {
	loanRepo    repository.LoanRepository
	nasabahRepo repository.NasabahRepository
	log         *zap.Logger
}

func NewLoanService(loanRepo repository.LoanRepository, nasabahRepo repository.NasabahRepository, log *zap.Logger) LoanService {
	return &loanService{loanRepo: loanRepo, nasabahRepo: nasabahRepo, log: log}
}

// ─── Apply ────────────────────────────────────────────────────────────────────

func (s *loanService) Apply(req models.CreateLoanRequest, userID uuid.UUID) (interface{}, error) {
	nasabah, err := s.nasabahRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("nasabah profile required before applying for a loan")
	}

	if security.IsSecure() {
		// OWASP API6 Secure: max pending loan applications per nasabah
		pending, err := s.loanRepo.CountPendingByNasabah(nasabah.ID)
		if err != nil {
			return nil, err
		}
		if pending >= MaxPendingLoansPerNasabah {
			return nil, fmt.Errorf("maximum %d pending loan applications allowed", MaxPendingLoansPerNasabah)
		}

		// OWASP API4 Secure: rate limit loan submissions per time window
		since := time.Now().Add(-time.Duration(LoanCreationWindowSec) * time.Second)
		recent, err := s.loanRepo.CountRecentByNasabah(nasabah.ID, since)
		if err != nil {
			return nil, err
		}
		if recent >= MaxLoansPerWindow {
			return nil, fmt.Errorf("too many loan applications — max %d per %ds", MaxLoansPerWindow, LoanCreationWindowSec)
		}

		// OWASP API3 Secure: validate amount range
		if req.Amount < 1_000_000 || req.Amount > 500_000_000 {
			return nil, errors.New("loan amount must be between Rp 1.000.000 and Rp 500.000.000")
		}
	}
	// TODO: Vulnerability Injection Point — OWASP API6 (Unrestricted Access to Sensitive Business Flows)
	// TODO: Vulnerability Injection Point — OWASP API4 (Unrestricted Resource Consumption)
	// Vulnerable: no pending limit, no rate limiting, no amount validation

	loan := &models.Loan{
		ID:           uuid.New(),
		NasabahID:    nasabah.ID,
		Amount:       req.Amount,
		InterestRate: req.InterestRate,
		TermMonths:   req.TermMonths,
		Status:       models.LoanStatusPending,
	}

	if err := s.loanRepo.Create(loan); err != nil {
		return nil, fmt.Errorf("create loan: %w", err)
	}

	s.log.Info("Loan application submitted",
		zap.String("loan_id", loan.ID.String()),
		zap.String("nasabah_id", nasabah.ID.String()),
		zap.Float64("amount", loan.Amount),
	)

	if security.IsVulnerable() {
		return loan.ToVulnerableResponse(), nil
	}
	return loan.ToResponse(), nil
}

// ─── GetByID ──────────────────────────────────────────────────────────────────

func (s *loanService) GetByID(loanID uuid.UUID, requestingUserID uuid.UUID, role string) (interface{}, error) {
	loan, err := s.loanRepo.FindByID(loanID)
	if err != nil {
		return nil, err
	}

	if security.IsSecure() && role == models.RoleNasabah {
		// OWASP API1 Secure: verify requester owns this loan via their nasabah profile
		nasabah, err := s.nasabahRepo.FindByUserID(requestingUserID)
		if err != nil || nasabah.ID != loan.NasabahID {
			s.log.Warn("BOLA attempt on loan",
				zap.String("requesting_user", requestingUserID.String()),
				zap.String("loan_id", loanID.String()),
			)
			// Return 404, not 403 — avoids confirming object existence to attacker
			return nil, repository.ErrNotFound
		}
	}
	// TODO: Vulnerability Injection Point — OWASP API1 (Broken Object Level Authorization)
	// Vulnerable: no ownership check — any authenticated user can fetch any loan by ID

	if security.IsVulnerable() {
		return loan.ToVulnerableResponse(), nil
	}
	return loan.ToResponse(), nil
}

// ─── List ─────────────────────────────────────────────────────────────────────

func (s *loanService) List(userID uuid.UUID, role string, page, limit int) (interface{}, error) {
	if page < 1 {
		page = defaultPage
	}

	if security.IsVulnerable() {
		// TODO: Vulnerability Injection Point — OWASP API4 (Unrestricted Resource Consumption)
		// TODO: Vulnerability Injection Point — OWASP API1 (BOLA)
		// Vulnerable: returns ALL loans regardless of requester's role
		s.log.Warn("[VULNERABLE] Loan list without ownership or pagination")
		loans, err := s.loanRepo.ListAll()
		if err != nil {
			return nil, err
		}
		result := make([]models.LoanVulnerableResponse, len(loans))
		for i, l := range loans {
			result[i] = l.ToVulnerableResponse()
		}
		return map[string]interface{}{
			"data":  result,
			"total": len(result),
			"note":  "[VULNERABLE] All loans returned — no BOLA check",
		}, nil
	}

	if limit < 1 || limit > maxSecureLimit {
		limit = defaultLimit
	}

	var (
		loans []models.Loan
		total int64
		err   error
	)

	switch role {
	case models.RoleAdmin, models.RoleStaff:
		// Admin/staff can list all loans
		loans, total, err = s.loanRepo.List(page, limit)
	default:
		// Nasabah: only own loans
		nasabah, lookupErr := s.nasabahRepo.FindByUserID(userID)
		if lookupErr != nil {
			return nil, errors.New("nasabah profile not found")
		}
		loans, total, err = s.loanRepo.FindByNasabahID(nasabah.ID, page, limit)
	}

	if err != nil {
		return nil, err
	}

	result := make([]models.LoanResponse, len(loans))
	for i, l := range loans {
		result[i] = l.ToResponse()
	}

	return map[string]interface{}{
		"data":  result,
		"total": total,
		"page":  page,
		"limit": limit,
	}, nil
}

// ─── Approve ──────────────────────────────────────────────────────────────────

func (s *loanService) Approve(loanID uuid.UUID, approverUserID uuid.UUID, role string) (interface{}, error) {
	loan, err := s.loanRepo.FindByID(loanID)
	if err != nil {
		return nil, err
	}

	if security.IsSecure() {
		// OWASP API5 Secure: only admin can approve loans
		if role != models.RoleAdmin {
			s.log.Warn("Unauthorized loan approval attempt",
				zap.String("user_id", approverUserID.String()),
				zap.String("role", role),
			)
			return nil, repository.ErrForbidden
		}
		if loan.Status != models.LoanStatusPending {
			return nil, fmt.Errorf("loan is not in pending status (current: %s)", loan.Status)
		}
	}
	// TODO: Vulnerability Injection Point — OWASP API5 (Broken Function Level Authorization)
	// Vulnerable: staff can approve loans — no role enforcement

	now := time.Now()
	loan.Status = models.LoanStatusApproved
	loan.ApprovedBy = &approverUserID
	loan.ApprovedAt = &now

	if err := s.loanRepo.Update(loan); err != nil {
		return nil, err
	}

	s.log.Info("Loan approved",
		zap.String("loan_id", loan.ID.String()),
		zap.String("approver", approverUserID.String()),
	)

	if security.IsVulnerable() {
		return loan.ToVulnerableResponse(), nil
	}
	return loan.ToResponse(), nil
}

// ─── Reject ───────────────────────────────────────────────────────────────────

func (s *loanService) Reject(loanID uuid.UUID, staffUserID uuid.UUID, role string, notes string) (interface{}, error) {
	loan, err := s.loanRepo.FindByID(loanID)
	if err != nil {
		return nil, err
	}

	if security.IsSecure() {
		// OWASP API5 Secure: admin or staff can reject
		if role == models.RoleNasabah {
			return nil, repository.ErrForbidden
		}
		if loan.Status != models.LoanStatusPending {
			return nil, fmt.Errorf("can only reject pending loans (current: %s)", loan.Status)
		}
	}

	loan.Status = models.LoanStatusRejected
	loan.Notes = notes

	if err := s.loanRepo.Update(loan); err != nil {
		return nil, err
	}

	s.log.Info("Loan rejected",
		zap.String("loan_id", loan.ID.String()),
		zap.String("by", staffUserID.String()),
	)

	if security.IsVulnerable() {
		return loan.ToVulnerableResponse(), nil
	}
	return loan.ToResponse(), nil
}

// ─── UpdateStatus (mass assignment demo) ─────────────────────────────────────

func (s *loanService) UpdateStatus(
	loanID uuid.UUID,
	req models.UpdateLoanStatusRequest,
	requestingUserID uuid.UUID,
	role string,
) (interface{}, error) {
	loan, err := s.loanRepo.FindByID(loanID)
	if err != nil {
		return nil, err
	}

	if security.IsSecure() {
		// OWASP API3 Secure: staff cannot set status to 'approved' — admin only
		if role == models.RoleStaff && req.Status == models.LoanStatusApproved {
			s.log.Warn("BOPLA: staff attempted to approve loan",
				zap.String("staff_id", requestingUserID.String()),
				zap.String("loan_id", loanID.String()),
			)
			return nil, errors.New("staff cannot approve loans — use /approve endpoint (admin only)")
		}
		// Nasabah cannot change status at all
		if role == models.RoleNasabah {
			return nil, repository.ErrForbidden
		}
	}
	// TODO: Vulnerability Injection Point — OWASP API3 (BOPLA / Mass Assignment)
	// Vulnerable: staff can set status to 'approved' — bypasses business rule

	loan.Status = req.Status
	if req.Notes != "" {
		loan.Notes = req.Notes
	}

	if err := s.loanRepo.Update(loan); err != nil {
		return nil, err
	}

	if security.IsVulnerable() {
		return loan.ToVulnerableResponse(), nil
	}
	return loan.ToResponse(), nil
}
