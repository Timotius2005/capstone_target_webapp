// Package mocks provides in-memory mock implementations of all repository interfaces.
// Each mock uses function fields so individual tests can override specific methods
// without implementing the entire interface.
package mocks

import (
	"time"

	"github.com/google/uuid"

	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/repository"
)

// ── UserRepository Mock ───────────────────────────────────────────────────────

type MockUserRepository struct {
	CreateFunc                 func(user *models.User) error
	FindByIDFunc               func(id uuid.UUID) (*models.User, error)
	FindByUsernameFunc         func(username string) (*models.User, error)
	FindByEmailFunc            func(email string) (*models.User, error)
	UpdateFunc                 func(user *models.User) error
	IncrementLoginAttemptsFunc func(id uuid.UUID) error
	ResetLoginAttemptsFunc     func(id uuid.UUID) error
	ListAllFunc                func(page, limit int) ([]models.User, int64, error)
}

func (m *MockUserRepository) Create(user *models.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return nil
}

func (m *MockUserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(id)
	}
	return nil, repository.ErrNotFound
}

func (m *MockUserRepository) FindByUsername(username string) (*models.User, error) {
	if m.FindByUsernameFunc != nil {
		return m.FindByUsernameFunc(username)
	}
	return nil, repository.ErrNotFound
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(email)
	}
	return nil, repository.ErrNotFound
}

func (m *MockUserRepository) Update(user *models.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(user)
	}
	return nil
}

func (m *MockUserRepository) IncrementLoginAttempts(id uuid.UUID) error {
	if m.IncrementLoginAttemptsFunc != nil {
		return m.IncrementLoginAttemptsFunc(id)
	}
	return nil
}

func (m *MockUserRepository) ResetLoginAttempts(id uuid.UUID) error {
	if m.ResetLoginAttemptsFunc != nil {
		return m.ResetLoginAttemptsFunc(id)
	}
	return nil
}

func (m *MockUserRepository) ListAll(page, limit int) ([]models.User, int64, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc(page, limit)
	}
	return nil, 0, nil
}

// ── NasabahRepository Mock ────────────────────────────────────────────────────

type MockNasabahRepository struct {
	CreateFunc      func(n *models.Nasabah) error
	FindByIDFunc    func(id uuid.UUID) (*models.Nasabah, error)
	FindByUserIDFunc func(userID uuid.UUID) (*models.Nasabah, error)
	FindByNIKFunc   func(nik string) (*models.Nasabah, error)
	UpdateFunc      func(n *models.Nasabah) error
	DeleteFunc      func(id uuid.UUID) error
	ListFunc        func(page, limit int) ([]models.Nasabah, int64, error)
	ListAllFunc     func() ([]models.Nasabah, error)
}

func (m *MockNasabahRepository) Create(n *models.Nasabah) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(n)
	}
	return nil
}

func (m *MockNasabahRepository) FindByID(id uuid.UUID) (*models.Nasabah, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(id)
	}
	return nil, repository.ErrNotFound
}

func (m *MockNasabahRepository) FindByUserID(userID uuid.UUID) (*models.Nasabah, error) {
	if m.FindByUserIDFunc != nil {
		return m.FindByUserIDFunc(userID)
	}
	return nil, repository.ErrNotFound
}

func (m *MockNasabahRepository) FindByNIK(nik string) (*models.Nasabah, error) {
	if m.FindByNIKFunc != nil {
		return m.FindByNIKFunc(nik)
	}
	return nil, repository.ErrNotFound
}

func (m *MockNasabahRepository) Update(n *models.Nasabah) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(n)
	}
	return nil
}

func (m *MockNasabahRepository) Delete(id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func (m *MockNasabahRepository) List(page, limit int) ([]models.Nasabah, int64, error) {
	if m.ListFunc != nil {
		return m.ListFunc(page, limit)
	}
	return nil, 0, nil
}

func (m *MockNasabahRepository) ListAll() ([]models.Nasabah, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc()
	}
	return nil, nil
}

// ── LoanRepository Mock ───────────────────────────────────────────────────────

type MockLoanRepository struct {
	CreateFunc                func(l *models.Loan) error
	FindByIDFunc              func(id uuid.UUID) (*models.Loan, error)
	FindByNasabahIDFunc       func(nasabahID uuid.UUID, page, limit int) ([]models.Loan, int64, error)
	ListFunc                  func(page, limit int) ([]models.Loan, int64, error)
	ListAllFunc               func() ([]models.Loan, error)
	UpdateFunc                func(l *models.Loan) error
	CountPendingByNasabahFunc func(nasabahID uuid.UUID) (int64, error)
	CountRecentByNasabahFunc  func(nasabahID uuid.UUID, since time.Time) (int64, error)
}

func (m *MockLoanRepository) Create(l *models.Loan) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(l)
	}
	return nil
}

func (m *MockLoanRepository) FindByID(id uuid.UUID) (*models.Loan, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(id)
	}
	return nil, repository.ErrNotFound
}

func (m *MockLoanRepository) FindByNasabahID(nasabahID uuid.UUID, page, limit int) ([]models.Loan, int64, error) {
	if m.FindByNasabahIDFunc != nil {
		return m.FindByNasabahIDFunc(nasabahID, page, limit)
	}
	return nil, 0, nil
}

func (m *MockLoanRepository) List(page, limit int) ([]models.Loan, int64, error) {
	if m.ListFunc != nil {
		return m.ListFunc(page, limit)
	}
	return nil, 0, nil
}

func (m *MockLoanRepository) ListAll() ([]models.Loan, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc()
	}
	return nil, nil
}

func (m *MockLoanRepository) Update(l *models.Loan) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(l)
	}
	return nil
}

func (m *MockLoanRepository) CountPendingByNasabah(nasabahID uuid.UUID) (int64, error) {
	if m.CountPendingByNasabahFunc != nil {
		return m.CountPendingByNasabahFunc(nasabahID)
	}
	return 0, nil
}

func (m *MockLoanRepository) CountRecentByNasabah(nasabahID uuid.UUID, since time.Time) (int64, error) {
	if m.CountRecentByNasabahFunc != nil {
		return m.CountRecentByNasabahFunc(nasabahID, since)
	}
	return 0, nil
}

// ── TransactionRepository Mock ────────────────────────────────────────────────

type MockTransactionRepository struct {
	CreateFunc        func(tx *models.Transaction) error
	FindByIDFunc      func(id uuid.UUID) (*models.Transaction, error)
	FindByLoanIDFunc  func(loanID uuid.UUID, page, limit int) ([]models.Transaction, int64, error)
	ListFunc          func(page, limit int) ([]models.Transaction, int64, error)
}

func (m *MockTransactionRepository) Create(tx *models.Transaction) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(tx)
	}
	return nil
}

func (m *MockTransactionRepository) FindByID(id uuid.UUID) (*models.Transaction, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(id)
	}
	return nil, repository.ErrNotFound
}

func (m *MockTransactionRepository) FindByLoanID(loanID uuid.UUID, page, limit int) ([]models.Transaction, int64, error) {
	if m.FindByLoanIDFunc != nil {
		return m.FindByLoanIDFunc(loanID, page, limit)
	}
	return nil, 0, nil
}

func (m *MockTransactionRepository) List(page, limit int) ([]models.Transaction, int64, error) {
	if m.ListFunc != nil {
		return m.ListFunc(page, limit)
	}
	return nil, 0, nil
}
