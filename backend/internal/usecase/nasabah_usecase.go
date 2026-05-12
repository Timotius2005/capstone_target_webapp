package usecase

import (
	"time"

	"github.com/google/uuid"
	"pt-dana-sejahtera/internal/domain"
	"pt-dana-sejahtera/internal/repository"
)

type NasabahUseCase struct {
	nasabahRepo *repository.NasabahRepository
}

func NewNasabahUseCase(nasabahRepo *repository.NasabahRepository) *NasabahUseCase {
	return &NasabahUseCase{nasabahRepo: nasabahRepo}
}

type CreateNasabahRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	NIK         string    `json:"nik"`
	Phone       string    `json:"phone"`
	Address     string    `json:"address"`
	DateOfBirth time.Time `json:"date_of_birth"`
}

func (uc *NasabahUseCase) GetAllNasabah() ([]*domain.Nasabah, error) {
	// TODO: SECURITY VULNERABILITY - No Authorization Check
	// Should check if user has permission to view all nasabah
	return uc.nasabahRepo.GetAll()
}

func (uc *NasabahUseCase) GetNasabahByID(id uuid.UUID) (*domain.Nasabah, error) {
	// TODO: SECURITY VULNERABILITY - Broken Object Level Authorization
	// Should check if user owns this nasabah record
	return uc.nasabahRepo.GetByID(id)
}

func (uc *NasabahUseCase) CreateNasabah(req CreateNasabahRequest) error {
	nasabah := &domain.Nasabah{
		ID:          uuid.New(),
		UserID:      req.UserID,
		Name:        req.Name,
		NIK:         req.NIK,
		Phone:       req.Phone,
		Address:     req.Address,
		DateOfBirth: req.DateOfBirth,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return uc.nasabahRepo.Create(nasabah)
}