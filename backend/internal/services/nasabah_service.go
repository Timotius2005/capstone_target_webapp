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

const (
	defaultPage    = 1
	defaultLimit   = 10
	maxSecureLimit = 100
)

type NasabahService interface {
	Create(req models.CreateNasabahRequest, userID uuid.UUID) (interface{}, error)
	GetByID(id uuid.UUID, requestingUserID uuid.UUID, role string) (interface{}, error)
	GetByUserID(userID uuid.UUID) (interface{}, error)
	List(page, limit int) (interface{}, error)
	Update(id uuid.UUID, req models.UpdateNasabahRequest, requestingUserID uuid.UUID, role string) (interface{}, error)
	Delete(id uuid.UUID, requestingUserID uuid.UUID, role string) error
}

type nasabahService struct {
	repo     repository.NasabahRepository
	userRepo repository.UserRepository
	log      *zap.Logger
}

func NewNasabahService(repo repository.NasabahRepository, userRepo repository.UserRepository, log *zap.Logger) NasabahService {
	return &nasabahService{repo: repo, userRepo: userRepo, log: log}
}

func (s *nasabahService) Create(req models.CreateNasabahRequest, userID uuid.UUID) (interface{}, error) {
	// Check if profile already registered for this user
	if _, err := s.repo.FindByUserID(userID); err == nil {
		return nil, errors.New("nasabah profile already registered for this user")
	}

	// A03 Secure: validate NIK uniqueness to prevent duplicate identities.
	if security.IsSecureFor(security.CategoryA03) {
		if _, err := s.repo.FindByNIK(req.NIK); err == nil {
			return nil, errors.New("NIK already registered")
		}
	}
	// TODO: Vulnerability Injection Point — OWASP API3 / A03 (Injection / BOPLA)
	// A03 enabled: no NIK uniqueness check — allows duplicate identities.

	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return nil, fmt.Errorf("invalid date_of_birth format, expected YYYY-MM-DD")
	}

	n := &models.Nasabah{
		ID:          uuid.New(),
		UserID:      userID,
		FullName:    req.FullName,
		NIK:         req.NIK,
		Phone:       req.Phone,
		Address:     req.Address,
		DateOfBirth: dob,
	}

	if err := s.repo.Create(n); err != nil {
		return nil, fmt.Errorf("create nasabah: %w", err)
	}

	s.log.Info("Nasabah profile created",
		zap.String("id", n.ID.String()),
		zap.String("user_id", userID.String()),
	)

	if security.IsVulnerableFor(security.CategoryA02) {
		// TODO: Vulnerability Injection Point — OWASP API3 / A02 (Cryptographic Failures)
		// A02 enabled: returns full NIK and internal fields in response.
		return n.ToVulnerableResponse(), nil
	}
	return n.ToResponse(), nil
}

func (s *nasabahService) GetByID(id uuid.UUID, requestingUserID uuid.UUID, role string) (interface{}, error) {
	n, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// A01 Secure: nasabah role can only access own data.
	if security.IsSecureFor(security.CategoryA01) {
		if role == models.RoleNasabah && n.UserID != requestingUserID {
			s.log.Warn("BOLA attempt on nasabah",
				zap.String("requester", requestingUserID.String()),
				zap.String("owner", n.UserID.String()),
			)
			return nil, repository.ErrForbidden
		}
	}
	// TODO: Vulnerability Injection Point — OWASP API1 / A01 (Broken Access Control / BOLA)
	// A01 enabled: no ownership check — any user can access any nasabah profile.

	if security.IsVulnerableFor(security.CategoryA02) {
		// TODO: Vulnerability Injection Point — OWASP API3 / A02 (Cryptographic Failures)
		// A02 enabled: response includes NIK, internal IDs, sensitive fields.
		return n.ToVulnerableResponse(), nil
	}
	return n.ToResponse(), nil
}

func (s *nasabahService) GetByUserID(userID uuid.UUID) (interface{}, error) {
	n, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if security.IsVulnerableFor(security.CategoryA02) {
		// TODO: Vulnerability Injection Point — A02 (Cryptographic Failures)
		// A02 enabled: sensitive fields (NIK) included in own-profile response.
		return n.ToVulnerableResponse(), nil
	}
	return n.ToResponse(), nil
}

func (s *nasabahService) List(page, limit int) (interface{}, error) {
	if page < 1 {
		page = defaultPage
	}

	if security.IsVulnerableFor(security.CategoryA04) {
		// TODO: Vulnerability Injection Point — OWASP API4 / A04 (Insecure Design)
		// A04 enabled: no pagination — full table dump including sensitive PII.
		s.log.Warn("[VULNERABLE] Nasabah list requested without pagination")
		list, err := s.repo.ListAll()
		if err != nil {
			return nil, err
		}
		result := make([]models.NasabahVulnerableResponse, len(list))
		for i, n := range list {
			result[i] = n.ToVulnerableResponse()
		}
		return map[string]interface{}{
			"data":  result,
			"total": len(result),
			"note":  "[VULNERABLE] No pagination — full NIK exposed",
		}, nil
	}

	if limit < 1 || limit > maxSecureLimit {
		limit = defaultLimit
	}

	list, total, err := s.repo.List(page, limit)
	if err != nil {
		return nil, err
	}

	result := make([]models.NasabahResponse, len(list))
	for i, n := range list {
		result[i] = n.ToResponse()
	}

	return map[string]interface{}{
		"data":  result,
		"total": total,
		"page":  page,
		"limit": limit,
	}, nil
}

func (s *nasabahService) Update(
	id uuid.UUID,
	req models.UpdateNasabahRequest,
	requestingUserID uuid.UUID,
	role string,
) (interface{}, error) {
	n, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// A01 Secure: nasabah can only update own profile.
	if security.IsSecureFor(security.CategoryA01) {
		if role == models.RoleNasabah && n.UserID != requestingUserID {
			return nil, repository.ErrForbidden
		}
	}
	// TODO: Vulnerability Injection Point — OWASP API1 / A01 (BOLA)
	// A01 enabled: any user can update any nasabah profile.

	if req.FullName != nil {
		n.FullName = *req.FullName
	}
	if req.Phone != nil {
		n.Phone = *req.Phone
	}
	if req.Address != nil {
		n.Address = *req.Address
	}

	if err := s.repo.Update(n); err != nil {
		return nil, err
	}

	if security.IsVulnerableFor(security.CategoryA02) {
		// TODO: Vulnerability Injection Point — A02 (Cryptographic Failures)
		// A02 enabled: sensitive fields returned in update response.
		return n.ToVulnerableResponse(), nil
	}
	return n.ToResponse(), nil
}

func (s *nasabahService) Delete(id uuid.UUID, requestingUserID uuid.UUID, role string) error {
	n, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// A01 Secure: only admin can delete nasabah profiles.
	if security.IsSecureFor(security.CategoryA01) && role != models.RoleAdmin {
		_ = n // suppress unused warning
		return repository.ErrForbidden
	}
	// TODO: Vulnerability Injection Point — OWASP API1 / A01 (BOLA)
	// A01 enabled: any role can delete any nasabah profile.

	return s.repo.Delete(id)
}
