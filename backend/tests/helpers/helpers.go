// Package helpers provides shared test utilities for the PT. Dana Sejahtera test suite.
// All helpers are pure (no side effects) and safe to call from parallel tests.
package helpers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/services"
)

// TestJWTSecret is the shared HMAC secret used in all test tokens.
// Must match the secret passed to services.NewAuthService in test setups.
const TestJWTSecret = "test-jwt-secret-for-unit-tests-32chars!!"

// MakeTestToken creates a valid, non-expired JWT for the given identity.
func MakeTestToken(userID, username, role string) string {
	claims := services.Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(TestJWTSecret))
	return token
}

// MakeExpiredToken creates an already-expired JWT — used to test OWASP API2.
func MakeExpiredToken(userID, username, role string) string {
	claims := services.Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(TestJWTSecret))
	return token
}

// MakeTamperedToken creates a JWT signed with a different key.
func MakeTamperedToken(userID, username, role string) string {
	claims := services.Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte("wrong-secret-key-for-tampering-attack!!"))
	return token
}

// MakeTestUser creates a test User with bcrypt-hashed password "TestPass123!".
func MakeTestUser(role string) *models.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte("TestPass123!"), bcrypt.MinCost)
	return &models.User{
		ID:           uuid.New(),
		Username:     "testuser_" + role,
		Email:        "test_" + role + "@danasejahtera.test",
		PasswordHash: string(hash),
		Role:         role,
		IsActive:     true,
		LoginAttempts: 0,
	}
}

// MakeLockedUser creates a test User whose LoginAttempts == 5 (locked).
func MakeLockedUser() *models.User {
	u := MakeTestUser(models.RoleNasabah)
	u.LoginAttempts = 5
	return u
}

// MakeTestNasabah creates a Nasabah profile belonging to the given userID.
func MakeTestNasabah(userID uuid.UUID) *models.Nasabah {
	return &models.Nasabah{
		ID:          uuid.New(),
		UserID:      userID,
		FullName:    "Budi Santoso",
		NIK:         "3201234567890001",
		Phone:       "+6281234567890",
		Address:     "Jl. Test No. 1, Jakarta Pusat",
		DateOfBirth: time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
	}
}

// MakeTestLoan creates a pending loan for the given nasabahID.
func MakeTestLoan(nasabahID uuid.UUID) *models.Loan {
	return &models.Loan{
		ID:           uuid.New(),
		NasabahID:    nasabahID,
		Amount:       10_000_000,
		InterestRate: 12.5,
		TermMonths:   24,
		Status:       models.LoanStatusPending,
	}
}

// ── HTTP Request Helpers ──────────────────────────────────────────────────────

// NewJSONRequest creates an HTTP request with JSON body and optional Bearer token.
func NewJSONRequest(method, path string, body interface{}, token string) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

// DoRequest performs req against router and returns the recorder.
func DoRequest(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ParseBody decodes JSON response body into dst.
func ParseBody(w *httptest.ResponseRecorder, dst interface{}) error {
	return json.Unmarshal(w.Body.Bytes(), dst)
}

// BearerHeader returns the Authorization header value.
func BearerHeader(token string) string {
	return "Bearer " + token
}
