package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"pt-dana-sejahtera/internal/models"
)

type LoanRepository interface {
	Create(l *models.Loan) error
	FindByID(id uuid.UUID) (*models.Loan, error)
	FindByNasabahID(nasabahID uuid.UUID, page, limit int) ([]models.Loan, int64, error)
	List(page, limit int) ([]models.Loan, int64, error)
	ListAll() ([]models.Loan, error) // vulnerable: no pagination
	Update(l *models.Loan) error
	CountPendingByNasabah(nasabahID uuid.UUID) (int64, error)
	CountRecentByNasabah(nasabahID uuid.UUID, since time.Time) (int64, error)
}

type loanRepository struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) LoanRepository {
	return &loanRepository{db: db}
}

func (r *loanRepository) Create(l *models.Loan) error {
	return r.db.Create(l).Error
}

func (r *loanRepository) FindByID(id uuid.UUID) (*models.Loan, error) {
	var l models.Loan
	if err := r.db.First(&l, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &l, nil
}

// FindByNasabahID — ownership-scoped query used in secure mode.
func (r *loanRepository) FindByNasabahID(nasabahID uuid.UUID, page, limit int) ([]models.Loan, int64, error) {
	var list []models.Loan
	var total int64
	offset := (page - 1) * limit

	q := r.db.Model(&models.Loan{}).Where("nasabah_id = ?", nasabahID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loanRepository) List(page, limit int) ([]models.Loan, int64, error) {
	var list []models.Loan
	var total int64
	offset := (page - 1) * limit

	if err := r.db.Model(&models.Loan{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListAll dumps entire loans table — vulnerable mode only.
// TODO: Vulnerability Injection Point — OWASP API4 (Unrestricted Resource Consumption)
func (r *loanRepository) ListAll() ([]models.Loan, error) {
	var list []models.Loan
	return list, r.db.Find(&list).Error
}

func (r *loanRepository) Update(l *models.Loan) error {
	return r.db.Save(l).Error
}

// CountPendingByNasabah checks for too many open applications.
// Used for OWASP API6 (Sensitive Business Flow) protection.
func (r *loanRepository) CountPendingByNasabah(nasabahID uuid.UUID) (int64, error) {
	var count int64
	return count, r.db.Model(&models.Loan{}).
		Where("nasabah_id = ? AND status = ?", nasabahID, models.LoanStatusPending).
		Count(&count).Error
}

// CountRecentByNasabah supports rate-limiting loan creation per time window.
func (r *loanRepository) CountRecentByNasabah(nasabahID uuid.UUID, since time.Time) (int64, error) {
	var count int64
	return count, r.db.Model(&models.Loan{}).
		Where("nasabah_id = ? AND created_at > ?", nasabahID, since).
		Count(&count).Error
}
