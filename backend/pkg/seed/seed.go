// Package seed inserts test users into the database for E2E and integration tests.
// It is safe to run multiple times — conflicts on username/email are silently skipped.
//
// Usage (from repo root):
//
//	cd backend && go run ./pkg/seed
package seed

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/repository"
)

// User describes a test user to be seeded.
type User struct {
	Username string
	Email    string
	Password string // plain-text; hashed with bcrypt cost 10 before insertion
	Role     string
}

// TestUsers lists every test account that E2E / integration tests depend on.
var TestUsers = []User{
	{
		Username: "budi_santoso",
		Email:    "budi_santoso@danasejahtera.id",
		Password: "Admin@Dana2025!",
		Role:     models.RoleAdmin,
	},
	{
		Username: "dewi_rahayu",
		Email:    "dewi_rahayu@danasejahtera.id",
		Password: "Staff@Dana2025!",
		Role:     models.RoleStaff,
	},
}

// Run upserts every entry in TestUsers into the database.
// Existing rows (matched by username or email) are left untouched.
func Run(db *gorm.DB) error {
	repo := repository.NewUserRepository(db)

	for _, u := range TestUsers {
		// Skip if already present.
		if _, err := repo.FindByUsername(u.Username); err == nil {
			continue
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
		if err != nil {
			return fmt.Errorf("bcrypt %s: %w", u.Username, err)
		}

		now := time.Now().UTC()
		user := &models.User{
			ID:           uuid.New(),
			Username:     u.Username,
			Email:        u.Email,
			PasswordHash: string(hash),
			Role:         u.Role,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := repo.Create(user); err != nil {
			// Race between check and insert — treat as benign.
			if !errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("create %s: %w", u.Username, err)
			}
		}
	}
	return nil
}
