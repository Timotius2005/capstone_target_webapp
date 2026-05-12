package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"pt-dana-sejahtera/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByUsername(username string) (*domain.User, error) {
	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE username = $1`
	row := r.db.QueryRow(query, username)

	var user domain.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var user domain.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Create(user *domain.User) error {
	query := `INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(query, user.ID, user.Username, user.Email, user.Password, user.Role, user.CreatedAt, user.UpdatedAt)
	return err
}