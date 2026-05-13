package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"pt-dana-sejahtera/internal/models"
)

type TransactionRepository interface {
	Create(t *models.Transaction) error
	FindByLoanID(loanID uuid.UUID, page, limit int) ([]models.Transaction, int64, error)
	List(page, limit int) ([]models.Transaction, int64, error)
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(t *models.Transaction) error {
	return r.db.Create(t).Error
}

func (r *transactionRepository) FindByLoanID(loanID uuid.UUID, page, limit int) ([]models.Transaction, int64, error) {
	var list []models.Transaction
	var total int64
	offset := (page - 1) * limit

	q := r.db.Model(&models.Transaction{}).Where("loan_id = ?", loanID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset(offset).Limit(limit).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *transactionRepository) List(page, limit int) ([]models.Transaction, int64, error) {
	var list []models.Transaction
	var total int64
	offset := (page - 1) * limit

	if err := r.db.Model(&models.Transaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
