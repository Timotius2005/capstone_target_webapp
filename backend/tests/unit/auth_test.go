package unit_test

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"pt-dana-sejahtera/internal/handlers"
	"pt-dana-sejahtera/internal/middleware"
	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/security"
	"pt-dana-sejahtera/internal/services"
	"pt-dana-sejahtera/tests/helpers"
	"pt-dana-sejahtera/tests/mocks"
)

func init() { gin.SetMode(gin.TestMode) }

// setupAuthRouter wires an auth handler with a mock user repo.
func setupAuthRouter(userRepo *mocks.MockUserRepository) *gin.Engine {
	log := zap.NewNop()
	authSvc := services.NewAuthService(userRepo, helpers.TestJWTSecret, log)
	authH := handlers.NewAuthHandler(authSvc, log)
	r := gin.New()
	r.POST("/auth/login", authH.Login)
	r.POST("/auth/register", authH.Register)
	r.GET("/auth/me", middleware.AuthRequired(authSvc), authH.Me)
	return r
}

// ── Register ──────────────────────────────────────────────────────────────────

func TestRegister_SecureMode_Success(t *testing.T) {
	security.SetMode(security.ModeSecure)
	userRepo := &mocks.MockUserRepository{
		FindByUsernameFunc: func(_ string) (*models.User, error) { return nil, errNotFound },
		FindByEmailFunc:    func(_ string) (*models.User, error) { return nil, errNotFound },
		CreateFunc:         func(_ *models.User) error { return nil },
	}
	w := helpers.DoRequest(setupAuthRouter(userRepo), helpers.NewJSONRequest("POST", "/auth/register",
		map[string]interface{}{"username": "newuser", "email": "newuser@test.com", "password": "SecurePass123!"},
		""))
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRegister_DuplicateUsername(t *testing.T) {
	security.SetMode(security.ModeSecure)
	existing := helpers.MakeTestUser(models.RoleNasabah)
	userRepo := &mocks.MockUserRepository{
		FindByUsernameFunc: func(_ string) (*models.User, error) { return existing, nil },
	}
	w := helpers.DoRequest(setupAuthRouter(userRepo), helpers.NewJSONRequest("POST", "/auth/register",
		map[string]interface{}{"username": existing.Username, "email": "other@test.com", "password": "SecurePass123!"},
		""))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_WeakPassword_Rejected(t *testing.T) {
	security.SetMode(security.ModeSecure)
	w := helpers.DoRequest(setupAuthRouter(&mocks.MockUserRepository{}),
		helpers.NewJSONRequest("POST", "/auth/register",
			map[string]interface{}{"username": "user1", "email": "u@t.com", "password": "short"},
			""))
	assert.Equal(t, http.StatusBadRequest, w.Code, "password shorter than 8 chars must be rejected")
}

// ── Login — Secure Mode ───────────────────────────────────────────────────────

func TestLogin_SecureMode_ValidCredentials(t *testing.T) {
	security.SetMode(security.ModeSecure)
	user := helpers.MakeTestUser(models.RoleNasabah)
	userRepo := &mocks.MockUserRepository{
		FindByUsernameFunc:     func(_ string) (*models.User, error) { return user, nil },
		ResetLoginAttemptsFunc: func(_ uuid.UUID) error { return nil },
		UpdateFunc:             func(_ *models.User) error { return nil },
	}
	w := helpers.DoRequest(setupAuthRouter(userRepo), helpers.NewJSONRequest("POST", "/auth/login",
		map[string]interface{}{"username": user.Username, "password": "TestPass123!"},
		""))

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, helpers.ParseBody(w, &body))
	assert.NotEmpty(t, body["token"])
	assert.NotEmpty(t, body["refresh_token"], "secure mode must include refresh_token")

	// OWASP API3: password_hash must be absent
	u, _ := body["user"].(map[string]interface{})
	assert.Nil(t, u["password_hash"], "secure mode must not expose password_hash")
	assert.Nil(t, u["login_attempts"], "secure mode must not expose login_attempts")
}

func TestLogin_SecureMode_WrongPassword_Generic_Error(t *testing.T) {
	// OWASP API2: secure mode returns generic error (not verbose)
	security.SetMode(security.ModeSecure)
	user := helpers.MakeTestUser(models.RoleNasabah)
	incrementCalled := false
	userRepo := &mocks.MockUserRepository{
		FindByUsernameFunc:         func(_ string) (*models.User, error) { return user, nil },
		IncrementLoginAttemptsFunc: func(_ uuid.UUID) error { incrementCalled = true; return nil },
	}
	w := helpers.DoRequest(setupAuthRouter(userRepo), helpers.NewJSONRequest("POST", "/auth/login",
		map[string]interface{}{"username": user.Username, "password": "WRONG"},
		""))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, incrementCalled, "secure mode must increment failed login counter")
	var body map[string]interface{}
	_ = helpers.ParseBody(w, &body)
	assert.Equal(t, "invalid credentials", body["error"],
		"secure mode must return generic error — not reveal username validity")
}

func TestLogin_SecureMode_AccountLocked_After_5_Attempts(t *testing.T) {
	// OWASP API2: account lockout prevents brute force in secure mode
	security.SetMode(security.ModeSecure)
	user := helpers.MakeLockedUser()
	userRepo := &mocks.MockUserRepository{
		FindByUsernameFunc: func(_ string) (*models.User, error) { return user, nil },
	}
	w := helpers.DoRequest(setupAuthRouter(userRepo), helpers.NewJSONRequest("POST", "/auth/login",
		map[string]interface{}{"username": user.Username, "password": "TestPass123!"},
		""))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var body map[string]interface{}
	_ = helpers.ParseBody(w, &body)
	assert.Contains(t, body["error"].(string), "locked")
}

// ── Login — Vulnerable Mode ───────────────────────────────────────────────────

func TestLogin_VulnerableMode_Plaintext_Compare(t *testing.T) {
	// OWASP API2: vulnerable mode uses == instead of bcrypt
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)

	user := helpers.MakeTestUser(models.RoleNasabah)
	user.PasswordHash = "plaintextpassword" // not a bcrypt hash

	userRepo := &mocks.MockUserRepository{
		FindByUsernameFunc: func(_ string) (*models.User, error) { return user, nil },
	}
	w := helpers.DoRequest(setupAuthRouter(userRepo), helpers.NewJSONRequest("POST", "/auth/login",
		map[string]interface{}{"username": user.Username, "password": "plaintextpassword"},
		""))

	require.Equal(t, http.StatusOK, w.Code, "vulnerable mode must accept plaintext password match")
	var body map[string]interface{}
	require.NoError(t, helpers.ParseBody(w, &body))
	u, _ := body["user"].(map[string]interface{})
	assert.NotNil(t, u["password_hash"], "OWASP API3: vulnerable mode must expose password_hash")
	assert.NotNil(t, u["login_attempts"], "OWASP API3: vulnerable mode must expose login_attempts")
}

func TestLogin_VulnerableMode_Verbose_Error_Reveals_Username(t *testing.T) {
	// OWASP API2: vulnerable mode error message leaks username existence
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)

	user := helpers.MakeTestUser(models.RoleNasabah)
	user.PasswordHash = "correctpassword"
	userRepo := &mocks.MockUserRepository{
		FindByUsernameFunc: func(_ string) (*models.User, error) { return user, nil },
	}
	w := helpers.DoRequest(setupAuthRouter(userRepo), helpers.NewJSONRequest("POST", "/auth/login",
		map[string]interface{}{"username": user.Username, "password": "wrongpassword"},
		""))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var body map[string]interface{}
	_ = helpers.ParseBody(w, &body)
	assert.Contains(t, body["error"].(string), user.Username,
		"OWASP API2: vulnerable mode must leak username in error response")
}

func TestLogin_VulnerableMode_No_Account_Lockout(t *testing.T) {
	// OWASP API2: vulnerable mode does not enforce lockout
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)

	user := helpers.MakeLockedUser()
	user.PasswordHash = "TestPass123!"
	userRepo := &mocks.MockUserRepository{
		FindByUsernameFunc: func(_ string) (*models.User, error) { return user, nil },
	}
	w := helpers.DoRequest(setupAuthRouter(userRepo), helpers.NewJSONRequest("POST", "/auth/login",
		map[string]interface{}{"username": user.Username, "password": "TestPass123!"},
		""))
	assert.Equal(t, http.StatusOK, w.Code,
		"OWASP API2: vulnerable mode must allow login even with 5 failed attempts")
}

// ── Token validation ──────────────────────────────────────────────────────────

func TestMe_ValidToken_Returns_200(t *testing.T) {
	security.SetMode(security.ModeSecure)
	user := helpers.MakeTestUser(models.RoleNasabah)
	token := helpers.MakeTestToken(user.ID.String(), user.Username, user.Role)
	w := helpers.DoRequest(setupAuthRouter(&mocks.MockUserRepository{}),
		helpers.NewJSONRequest("GET", "/auth/me", nil, token))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMe_ExpiredToken_Returns_401(t *testing.T) {
	security.SetMode(security.ModeSecure)
	user := helpers.MakeTestUser(models.RoleNasabah)
	expired := helpers.MakeExpiredToken(user.ID.String(), user.Username, user.Role)
	w := helpers.DoRequest(setupAuthRouter(&mocks.MockUserRepository{}),
		helpers.NewJSONRequest("GET", "/auth/me", nil, expired))
	assert.Equal(t, http.StatusUnauthorized, w.Code, "expired token must be rejected")
}

func TestMe_TamperedToken_Returns_401(t *testing.T) {
	security.SetMode(security.ModeSecure)
	user := helpers.MakeTestUser(models.RoleNasabah)
	tampered := helpers.MakeTamperedToken(user.ID.String(), user.Username, user.Role)
	w := helpers.DoRequest(setupAuthRouter(&mocks.MockUserRepository{}),
		helpers.NewJSONRequest("GET", "/auth/me", nil, tampered))
	assert.Equal(t, http.StatusUnauthorized, w.Code, "tampered token must be rejected")
}

func TestMe_NoToken_Returns_401(t *testing.T) {
	security.SetMode(security.ModeSecure)
	w := helpers.DoRequest(setupAuthRouter(&mocks.MockUserRepository{}),
		helpers.NewJSONRequest("GET", "/auth/me", nil, ""))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ── sentinel ──────────────────────────────────────────────────────────────────

// errNotFound satisfies the error interface for mock "not found" returns.
// Kept local to avoid importing repository package in every test.
type notFoundError struct{}

func (notFoundError) Error() string { return "record not found" }

var errNotFound error = notFoundError{}
