package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"pt-dana-sejahtera/internal/models"
)

type NasabahRepository interface {
	Create(n *models.Nasabah) error
	FindByID(id uuid.UUID) (*models.Nasabah, error)
	FindByUserID(userID uuid.UUID) (*models.Nasabah, error)
	FindByNIK(nik string) (*models.Nasabah, error)
	Update(n *models.Nasabah) error
	Delete(id uuid.UUID) error
	List(page, limit int) ([]models.Nasabah, int64, error)
	ListAll() ([]models.Nasabah, error) // vulnerable: no pagination
}

type nasabahRepository struct {
	db *gorm.DB
}

func NewNasabahRepository(db *gorm.DB) NasabahRepository {
	return &nasabahRepository{db: db}
}

func (r *nasabahRepository) Create(n *models.Nasabah) error {
	return r.db.Create(n).Error
}

func (r *nasabahRepository) FindByID(id uuid.UUID) (*models.Nasabah, error) {
	var n models.Nasabah
	if err := r.db.First(&n, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &n, nil
}

func (r *nasabahRepository) FindByUserID(userID uuid.UUID) (*models.Nasabah, error) {
	var n models.Nasabah
	if err := r.db.First(&n, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &n, nil
}

func (r *nasabahRepository) FindByNIK(nik string) (*models.Nasabah, error) {
	var n models.Nasabah
	if err := r.db.First(&n, "nik = ?", nik).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &n, nil
}

func (r *nasabahRepository) Update(n *models.Nasabah) error {
	return r.db.Save(n).Error
}

func (r *nasabahRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Nasabah{}, "id = ?", id).Error
}

func (r *nasabahRepository) List(page, limit int) ([]models.Nasabah, int64, error) {
	var list []models.Nasabah
	var total int64
	offset := (page - 1) * limit

	if err := r.db.Model(&models.Nasabah{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListAll dumps entire table — used only in vulnerable mode.
// TODO: Vulnerability Injection Point — OWASP API4 (Unrestricted Resource Consumption)
func (r *nasabahRepository) ListAll() ([]models.Nasabah, error) {
	var list []models.Nasabah
	return list, r.db.Find(&list).Error
}
