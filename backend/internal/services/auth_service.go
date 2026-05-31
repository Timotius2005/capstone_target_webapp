package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/repository"
	"pt-dana-sejahtera/internal/security"
)

// ─── Request / Response ───────────────────────────────────────────────────────

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50,alphanum"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role"     binding:"omitempty,oneof=admin staff nasabah"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token        string             `json:"token"`
	RefreshToken string             `json:"refresh_token,omitempty"`
	ExpiresAt    *time.Time         `json:"expires_at,omitempty"`
	User         models.UserResponse `json:"user"`
}

// LoginVulnerableResponse exposes the full user object including internals.
// TODO: Vulnerability Injection Point — OWASP API3 (BOPLA)
type LoginVulnerableResponse struct {
	Token string                      `json:"token"`
	User  models.UserVulnerableResponse `json:"user"`
}

// ─── JWT Claims ───────────────────────────────────────────────────────────────

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// ─── Service ──────────────────────────────────────────────────────────────────

type RegisterResponse struct {
	Token string              `json:"token"`
	User  models.UserResponse `json:"user"`
}

type AuthService interface {
	Register(req RegisterRequest) (*RegisterResponse, error)
	Login(req LoginRequest) (interface{}, error)
	ValidateToken(tokenStr string) (*Claims, error)
	RefreshToken(refreshToken string) (string, error)
}

type authService struct {
	userRepo  repository.UserRepository
	jwtSecret string
	log       *zap.Logger
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, log *zap.Logger) AuthService {
	return &authService{userRepo: userRepo, jwtSecret: jwtSecret, log: log}
}

// ─── Register ─────────────────────────────────────────────────────────────────

func (s *authService) Register(req RegisterRequest) (*RegisterResponse, error) {
	if _, err := s.userRepo.FindByUsername(req.Username); err == nil {
		return nil, errors.New("username already taken")
	}
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	role := models.RoleNasabah
	if req.Role != "" {
		role = req.Role
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Issue a token directly — bypasses login so it works in both secure and vulnerable mode.
	token, err := s.generateToken(user, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	s.log.Info("User registered",
		zap.String("username", user.Username),
		zap.String("role", user.Role),
		zap.String("id", user.ID.String()),
	)
	return &RegisterResponse{Token: token, User: user.ToResponse()}, nil
}

// ─── Login ────────────────────────────────────────────────────────────────────

func (s *authService) Login(req LoginRequest) (interface{}, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if security.IsVulnerableFor(security.CategoryA07) {
		// TODO: Vulnerability Injection Point — OWASP API2 / A07 (Authentication Failures)
		// A07 enabled: plain-text string comparison attempted first — no bcrypt, vulnerable
		// to timing attacks. Falls back to bcrypt so seeded users can still login.
		// JWT has no expiry (100-year token). Verbose error reveals username existence.
		plainMatch := user.PasswordHash == req.Password
		bcryptMatch := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) == nil
		if !plainMatch && !bcryptMatch {
			s.log.Warn("[VULNERABLE] Login failed — plain-text compare",
				zap.String("username", req.Username),
			)
			// TODO: Vulnerability Injection Point — OWASP API2 / A07
			// Detailed error message reveals whether username exists.
			return nil, fmt.Errorf("invalid credentials: password mismatch for user '%s'", req.Username)
		}

		// No-expiry token (100 years).
		token, err := s.generateToken(user, 100*365*24*time.Hour)
		if err != nil {
			return nil, err
		}

		s.log.Warn("[VULNERABLE] JWT issued with no expiry",
			zap.String("user", user.Username),
			zap.String("token_preview", token[:min(len(token), 20)]+"..."),
		)

		return LoginVulnerableResponse{
			Token: token,
			User:  user.ToVulnerableResponse(),
		}, nil
	}

	// ── Secure path ───────────────────────────────────────────────────────────

	// OWASP API2 Secure: account lockout after 5 failed attempts.
	if user.LoginAttempts >= 5 {
		return nil, errors.New("account locked — contact support after 5 failed attempts")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		_ = s.userRepo.IncrementLoginAttempts(user.ID)
		s.log.Warn("Failed login attempt",
			zap.String("username", req.Username),
			zap.Int("attempts", user.LoginAttempts+1),
		)
		return nil, errors.New("invalid credentials")
	}

	_ = s.userRepo.ResetLoginAttempts(user.ID)
	now := time.Now()
	user.LastLoginAt = &now
	_ = s.userRepo.Update(user)

	expiry := 15 * time.Minute
	expiresAt := now.Add(expiry)

	token, err := s.generateToken(user, expiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	s.log.Info("Login successful",
		zap.String("user", user.Username),
		zap.String("role", user.Role),
	)

	return LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    &expiresAt,
		User:         user.ToResponse(),
	}, nil
}

// ─── Token helpers ────────────────────────────────────────────────────────────

func (s *authService) generateToken(user *models.User, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
}

func (s *authService) generateRefreshToken(user *models.User) (string, error) {
	claims := Claims{
		UserID: user.ID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(s.jwtSecret + "-refresh"))
}

func (s *authService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func (s *authService) RefreshToken(refreshToken string) (string, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret + "-refresh"), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", errors.New("invalid refresh token claims")
	}

	user, err := s.userRepo.FindByID(uuid.MustParse(claims.UserID))
	if err != nil {
		return "", errors.New("user not found")
	}

	return s.generateToken(user, 15*time.Minute)
}

