package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"pt-dana-sejahtera/internal/domain"
)

type NasabahRepository struct {
	db *sql.DB
}

func NewNasabahRepository(db *sql.DB) *NasabahRepository {
	return &NasabahRepository{db: db}
}

func (r *NasabahRepository) GetAll() ([]*domain.Nasabah, error) {
	query := `SELECT id, user_id, name, nik, phone, address, date_of_birth, created_at, updated_at FROM nasabah`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nasabah []*domain.Nasabah
	for rows.Next() {
		var n domain.Nasabah
		err := rows.Scan(&n.ID, &n.UserID, &n.Name, &n.NIK, &n.Phone, &n.Address, &n.DateOfBirth, &n.CreatedAt, &n.UpdatedAt)
		if err != nil {
			return nil, err
		}
		nasabah = append(nasabah, &n)
	}

	return nasabah, nil
}

func (r *NasabahRepository) GetByID(id uuid.UUID) (*domain.Nasabah, error) {
	query := `SELECT id, user_id, name, nik, phone, address, date_of_birth, created_at, updated_at FROM nasabah WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var n domain.Nasabah
	err := row.Scan(&n.ID, &n.UserID, &n.Name, &n.NIK, &n.Phone, &n.Address, &n.DateOfBirth, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func (r *NasabahRepository) Create(nasabah *domain.Nasabah) error {
	query := `INSERT INTO nasabah (id, user_id, name, nik, phone, address, date_of_birth, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(query, nasabah.ID, nasabah.UserID, nasabah.Name, nasabah.NIK, nasabah.Phone, nasabah.Address, nasabah.DateOfBirth, nasabah.CreatedAt, nasabah.UpdatedAt)
	return err
}