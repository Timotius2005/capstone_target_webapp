package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"pt-dana-sejahtera/internal/domain"
	"pt-dana-sejahtera/internal/repository"
)

type AuthUseCase struct {
	userRepo *repository.UserRepository
}

func NewAuthUseCase(userRepo *repository.UserRepository) *AuthUseCase {
	return &AuthUseCase{userRepo: userRepo}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

func (uc *AuthUseCase) Login(req LoginRequest) (*LoginResponse, error) {
	user, err := uc.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// TODO: SECURITY VULNERABILITY - No Password Hashing Comparison
	// This should use bcrypt.CompareHashAndPassword
	if req.Password != user.Password {
		return nil, errors.New("invalid credentials")
	}

	// TODO: SECURITY VULNERABILITY - JWT Token Generation
	// Should generate proper JWT token with claims
	token := "fake-jwt-token-" + user.ID.String()

	return &LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (uc *AuthUseCase) Register(username, email, password string) error {
	// TODO: SECURITY VULNERABILITY - Weak Password Requirements
	// Should validate password strength

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &domain.User{
		ID:        uuid.New(),
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return uc.userRepo.Create(user)
}