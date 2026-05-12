package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"pt-dana-sejahtera/internal/domain"
)

type LoanRepository struct {
	db *sql.DB
}

func NewLoanRepository(db *sql.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

func (r *LoanRepository) GetAll() ([]*domain.Loan, error) {
	query := `SELECT id, nasabah_id, amount, interest_rate, term_months, status, approved_at, created_at, updated_at FROM loans`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []*domain.Loan
	for rows.Next() {
		var l domain.Loan
		err := rows.Scan(&l.ID, &l.NasabahID, &l.Amount, &l.InterestRate, &l.TermMonths, &l.Status, &l.ApprovedAt, &l.CreatedAt, &l.UpdatedAt)
		if err != nil {
			return nil, err
		}
		loans = append(loans, &l)
	}

	return loans, nil
}

func (r *LoanRepository) GetByID(id uuid.UUID) (*domain.Loan, error) {
	query := `SELECT id, nasabah_id, amount, interest_rate, term_months, status, approved_at, created_at, updated_at FROM loans WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var l domain.Loan
	err := row.Scan(&l.ID, &l.NasabahID, &l.Amount, &l.InterestRate, &l.TermMonths, &l.Status, &l.ApprovedAt, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &l, nil
}

func (r *LoanRepository) Create(loan *domain.Loan) error {
	query := `INSERT INTO loans (id, nasabah_id, amount, interest_rate, term_months, status, approved_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(query, loan.ID, loan.NasabahID, loan.Amount, loan.InterestRate, loan.TermMonths, loan.Status, loan.ApprovedAt, loan.CreatedAt, loan.UpdatedAt)
	return err
}