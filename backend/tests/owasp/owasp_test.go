// Package owasp_test validates all 10 OWASP API Security Top 10 categories
// against BOTH secure and vulnerable modes of PT. Dana Sejahtera backend.
//
// Each test documents:
//   - Which OWASP category it covers
//   - Expected behaviour in secure mode (protection active)
//   - Expected behaviour in vulnerable mode (intentional weakness)
//
// Exit code 0 only when ALL secure-mode assertions pass.
package owasp_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"pt-dana-sejahtera/internal/handlers"
	"pt-dana-sejahtera/internal/middleware"
	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/repository"
	"pt-dana-sejahtera/internal/security"
	"pt-dana-sejahtera/internal/services"
	"pt-dana-sejahtera/tests/helpers"
	"pt-dana-sejahtera/tests/mocks"
)

func init() { gin.SetMode(gin.TestMode) }

// ── Test app wiring ───────────────────────────────────────────────────────────

type testApp struct {
	router      *gin.Engine
	userRepo    *mocks.MockUserRepository
	nasabahRepo *mocks.MockNasabahRepository
	loanRepo    *mocks.MockLoanRepository
	txRepo      *mocks.MockTransactionRepository
	authSvc     services.AuthService
}

func newTestApp() *testApp {
	log := zap.NewNop()
	userRepo := &mocks.MockUserRepository{}
	nasabahRepo := &mocks.MockNasabahRepository{}
	loanRepo := &mocks.MockLoanRepository{}
	txRepo := &mocks.MockTransactionRepository{}

	authSvc := services.NewAuthService(userRepo, helpers.TestJWTSecret, log)
	nasabahSvc := services.NewNasabahService(nasabahRepo, userRepo, log)
	loanSvc := services.NewLoanService(loanRepo, nasabahRepo, log)
	txSvc := services.NewTransactionService(txRepo, loanRepo, nasabahRepo, log)
	extSvc := services.NewExternalService(log)

	authH := handlers.NewAuthHandler(authSvc, log)
	nasabahH := handlers.NewNasabahHandler(nasabahSvc, log)
	loanH := handlers.NewLoanHandler(loanSvc, log)
	_ = handlers.NewTransactionHandler(txSvc, log) // wired but not directly tested here
	adminH := handlers.NewAdminHandler(userRepo, log)
	ssrfH := handlers.NewSSRFHandler(extSvc, log)
	configH := handlers.NewConfigHandler(log)
	systemH := handlers.NewSystemHandler(nil, log) // nil DB: no persistence in tests

	r := gin.New()
	r.Use(middleware.CORS())
	r.Use(middleware.SecureHeaders())

	// Public
	r.GET("/config/mode", configH.GetMode)
	r.GET("/api/system/mode", systemH.GetMode)
	r.PUT("/api/system/mode", systemH.SetMode)

	auth := r.Group("/api/v1/auth")
	auth.POST("/register", authH.Register)
	auth.POST("/login", authH.Login)

	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthRequired(authSvc))
	protected.GET("/auth/me", authH.Me)

	nas := protected.Group("/nasabah")
	nas.POST("", nasabahH.Register)
	nas.GET("/me", nasabahH.GetMyProfile)
	nas.GET("", nasabahH.List)
	nas.GET("/:id", nasabahH.GetByID)

	loans := protected.Group("/loans")
	loans.POST("", loanH.Apply)
	loans.GET("", loanH.List)
	loans.GET("/:id", loanH.GetByID)
	loans.PATCH("/:id/status", loanH.UpdateStatus)
	loans.POST("/:id/approve", loanH.Approve)

	admin := protected.Group("/admin")
	admin.Use(middleware.AdminOnly())
	admin.GET("/users", adminH.ListUsers)
	admin.PUT("/users/:id/role", adminH.UpdateRole)

	protected.POST("/internal/fetch", ssrfH.Fetch)

	// Vulnerable-only deprecated routes (registered regardless for testing)
	v0 := r.Group("/api/v0")
	v0.GET("/loans", loanH.ListPublic)
	v0.GET("/users", adminH.ListUsersPublic)
	v0.GET("/debug", adminH.Debug)

	return &testApp{
		router: r, userRepo: userRepo, nasabahRepo: nasabahRepo,
		loanRepo: loanRepo, txRepo: txRepo, authSvc: authSvc,
	}
}

func do(app *testApp, req *http.Request) *httptest.ResponseRecorder {
	return helpers.DoRequest(app.router, req)
}

func parseBody(w *httptest.ResponseRecorder) map[string]interface{} {
	var m map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &m)
	return m
}

// ─────────────────────────────────────────────────────────────────────────────
// OWASP API1 — Broken Object Level Authorization (BOLA / IDOR)
// ─────────────────────────────────────────────────────────────────────────────

func TestAPI1_BOLA_SecureMode_Blocks_Cross_User_Access(t *testing.T) {
	security.SetMode(security.ModeSecure)
	app := newTestApp()

	ownerID := uuid.New()
	attackerID := uuid.New()
	nasabah := helpers.MakeTestNasabah(ownerID) // belongs to owner

	app.nasabahRepo.FindByIDFunc = func(_ uuid.UUID) (*models.Nasabah, error) {
		return nasabah, nil
	}

	// Attacker (different user, nasabah role) tries to access owner's nasabah.
	// The handler intentionally maps ErrForbidden → 404 to hide resource existence
	// (prevents enumeration even when ownership check fails).
	attackerToken := helpers.MakeTestToken(attackerID.String(), "attacker", models.RoleNasabah)
	w := do(app, helpers.NewJSONRequest("GET", "/api/v1/nasabah/"+nasabah.ID.String(), nil, attackerToken))

	assert.Equal(t, http.StatusNotFound, w.Code,
		"OWASP API1 Secure: cross-user access must be denied (404 hides resource existence)")
}

func TestAPI1_BOLA_VulnerableMode_Allows_Cross_User_Access(t *testing.T) {
	// OWASP API1 Vulnerable: no ownership check — IDOR is possible
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)
	app := newTestApp()

	ownerID := uuid.New()
	attackerID := uuid.New()
	nasabah := helpers.MakeTestNasabah(ownerID)

	app.nasabahRepo.FindByIDFunc = func(_ uuid.UUID) (*models.Nasabah, error) {
		return nasabah, nil
	}

	attackerToken := helpers.MakeTestToken(attackerID.String(), "attacker", models.RoleNasabah)
	w := do(app, helpers.NewJSONRequest("GET", "/api/v1/nasabah/"+nasabah.ID.String(), nil, attackerToken))

	assert.Equal(t, http.StatusOK, w.Code,
		"OWASP API1 Vulnerable: cross-user IDOR must be allowed (intentional weakness)")
}

func TestAPI1_BOLA_Admin_Always_Allowed(t *testing.T) {
	security.SetMode(security.ModeSecure)
	app := newTestApp()

	ownerID := uuid.New()
	adminID := uuid.New()
	nasabah := helpers.MakeTestNasabah(ownerID)
	app.nasabahRepo.FindByIDFunc = func(_ uuid.UUID) (*models.Nasabah, error) { return nasabah, nil }

	adminToken := helpers.MakeTestToken(adminID.String(), "budi_santoso", models.RoleAdmin)
	w := do(app, helpers.NewJSONRequest("GET", "/api/v1/nasabah/"+nasabah.ID.String(), nil, adminToken))
	assert.Equal(t, http.StatusOK, w.Code, "admin must be able to access any nasabah in secure mode")
}

// ─────────────────────────────────────────────────────────────────────────────
// OWASP API2 — Broken Authentication
// ─────────────────────────────────────────────────────────────────────────────

func TestAPI2_Auth_SecureMode_Bcrypt_Required(t *testing.T) {
	// Secure: bcrypt hash comparison — plaintext "password" won't match a bcrypt hash
	security.SetMode(security.ModeSecure)
	app := newTestApp()

	user := helpers.MakeTestUser(models.RoleNasabah) // has bcrypt hash
	app.userRepo.FindByUsernameFunc = func(_ string) (*models.User, error) { return user, nil }

	// Send the raw plaintext — it won't pass bcrypt.Compare even if correct
	w := do(app, helpers.NewJSONRequest("POST", "/api/v1/auth/login",
		map[string]interface{}{"username": user.Username, "password": "plaintexttest"}, ""))
	assert.Equal(t, http.StatusUnauthorized, w.Code,
		"OWASP API2 Secure: bcrypt comparison must reject plaintext that doesn't match hash")
}

func TestAPI2_Auth_SecureMode_JWT_Has_Expiry(t *testing.T) {
	security.SetMode(security.ModeSecure)
	user := helpers.MakeTestUser(models.RoleNasabah)
	app := newTestApp()
	app.userRepo.FindByUsernameFunc = func(_ string) (*models.User, error) { return user, nil }
	app.userRepo.ResetLoginAttemptsFunc = func(_ uuid.UUID) error { return nil }
	app.userRepo.UpdateFunc = func(_ *models.User) error { return nil }

	w := do(app, helpers.NewJSONRequest("POST", "/api/v1/auth/login",
		map[string]interface{}{"username": user.Username, "password": "TestPass123!"}, ""))
	require.Equal(t, http.StatusOK, w.Code)

	body := parseBody(w)
	assert.NotNil(t, body["expires_at"], "OWASP API2 Secure: JWT must include expiry time")
}

func TestAPI2_Auth_NoToken_Returns_401(t *testing.T) {
	security.SetMode(security.ModeSecure)
	app := newTestApp()
	w := do(app, helpers.NewJSONRequest("GET", "/api/v1/auth/me", nil, ""))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPI2_Auth_BothModes_Auth_Always_Required(t *testing.T) {
	for _, mode := range []security.ModeValue{security.ModeSecure, security.ModeSandbox} {
		security.SetMode(mode)
		app := newTestApp()
		w := do(app, helpers.NewJSONRequest("GET", "/api/v1/auth/me", nil, ""))
		assert.Equal(t, http.StatusUnauthorized, w.Code,
			"auth must be enforced in mode=%s", mode)
	}
	security.SetMode(security.ModeSecure)
}

// ─────────────────────────────────────────────────────────────────────────────
// OWASP API3 — Broken Object Property Level Authorization (BOPLA)
// ─────────────────────────────────────────────────────────────────────────────

func TestAPI3_BOPLA_SecureMode_NIK_Is_Masked(t *testing.T) {
	// Secure: NIK returned as "XXXX••••••••XXXX"
	security.SetMode(security.ModeSecure)
	app := newTestApp()

	userID := uuid.New()
	nasabah := helpers.MakeTestNasabah(userID)
	app.nasabahRepo.FindByUserIDFunc = func(_ uuid.UUID) (*models.Nasabah, error) { return nasabah, nil }

	token := helpers.MakeTestToken(userID.String(), "user", models.RoleNasabah)
	w := do(app, helpers.NewJSONRequest("GET", "/api/v1/nasabah/me", nil, token))
	require.Equal(t, http.StatusOK, w.Code)

	body := parseBody(w)
	nik, _ := body["nik"].(string)
	assert.Contains(t, nik, "••", "OWASP API3 Secure: NIK must be masked in response")
	assert.NotEqual(t, nasabah.NIK, nik, "full NIK must not be returned in secure mode")
}

func TestAPI3_BOPLA_VulnerableMode_NIK_Exposed(t *testing.T) {
	// OWASP API3 Vulnerable: full 16-digit NIK in response
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)
	app := newTestApp()

	userID := uuid.New()
	nasabah := helpers.MakeTestNasabah(userID)
	app.nasabahRepo.FindByUserIDFunc = func(_ uuid.UUID) (*models.Nasabah, error) { return nasabah, nil }

	token := helpers.MakeTestToken(userID.String(), "user", models.RoleNasabah)
	w := do(app, helpers.NewJSONRequest("GET", "/api/v1/nasabah/me", nil, token))
	require.Equal(t, http.StatusOK, w.Code)

	body := parseBody(w)
	nik, _ := body["nik"].(string)
	assert.Equal(t, nasabah.NIK, nik,
		"OWASP API3 Vulnerable: full NIK must be exposed (intentional weakness)")
}

func TestAPI3_BOPLA_Login_SecureMode_No_PasswordHash(t *testing.T) {
	security.SetMode(security.ModeSecure)
	app := newTestApp()
	user := helpers.MakeTestUser(models.RoleAdmin)
	app.userRepo.FindByUsernameFunc = func(_ string) (*models.User, error) { return user, nil }
	app.userRepo.ResetLoginAttemptsFunc = func(_ uuid.UUID) error { return nil }
	app.userRepo.UpdateFunc = func(_ *models.User) error { return nil }

	w := do(app, helpers.NewJSONRequest("POST", "/api/v1/auth/login",
		map[string]interface{}{"username": user.Username, "password": "TestPass123!"}, ""))
	require.Equal(t, http.StatusOK, w.Code)
	body := parseBody(w)
	u, _ := body["user"].(map[string]interface{})
	assert.Nil(t, u["password_hash"], "OWASP API3 Secure: password_hash must be absent")
}

// ─────────────────────────────────────────────────────────────────────────────
// OWASP API4 — Unrestricted Resource Consumption
// ─────────────────────────────────────────────────────────────────────────────

func TestAPI4_ResourceConsumption_SecureMode_List_Is_Paginated(t *testing.T) {
	// Secure: pagination is enforced; cannot dump full table
	security.SetMode(security.ModeSecure)
	app := newTestApp()

	// Repo returns a large list simulating full table
	bigList := make([]models.Nasabah, 500)
	for i := range bigList {
		bigList[i] = *helpers.MakeTestNasabah(uuid.New())
	}
	app.nasabahRepo.ListFunc = func(page, limit int) ([]models.Nasabah, int64, error) {
		// Should only return up to maxSecureLimit (100)
		end := limit
		if end > len(bigList) {
			end = len(bigList)
		}
		return bigList[:end], int64(len(bigList)), nil
	}

	adminToken := helpers.MakeTestToken(uuid.New().String(), "budi_santoso", models.RoleAdmin)
	w := do(app, helpers.NewJSONRequest("GET", "/api/v1/nasabah?page=1&limit=1000", nil, adminToken))
	require.Equal(t, http.StatusOK, w.Code)

	body := parseBody(w)
	data, _ := body["data"].([]interface{})
	assert.LessOrEqual(t, len(data), 100,
		"OWASP API4 Secure: list response must be capped at maxSecureLimit (100)")
}

func TestAPI4_ResourceConsumption_VulnerableMode_Returns_All(t *testing.T) {
	// OWASP API4 Vulnerable: no pagination — full table dump
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)
	app := newTestApp()

	bigList := make([]models.Nasabah, 500)
	for i := range bigList {
		bigList[i] = *helpers.MakeTestNasabah(uuid.New())
	}
	app.nasabahRepo.ListAllFunc = func() ([]models.Nasabah, error) { return bigList, nil }

	adminToken := helpers.MakeTestToken(uuid.New().String(), "budi_santoso", models.RoleAdmin)
	w := do(app, helpers.NewJSONRequest("GET", "/api/v1/nasabah", nil, adminToken))
	require.Equal(t, http.StatusOK, w.Code)

	body := parseBody(w)
	// Vulnerable response uses "note" field indicating no pagination
	assert.NotNil(t, body["note"], "OWASP API4 Vulnerable: response must contain warning note")
}

// ─────────────────────────────────────────────────────────────────────────────
// OWASP API5 — Broken Function Level Authorization (BFLA)
// ─────────────────────────────────────────────────────────────────────────────

func TestAPI5_BFLA_SecureMode_Staff_Cannot_Access_Admin_Routes(t *testing.T) {
	security.SetMode(security.ModeSecure)
	app := newTestApp()
	staffToken := helpers.MakeTestToken(uuid.New().String(), "dewi_rahayu", models.RoleStaff)
	w := do(app, helpers.NewJSONRequest("GET", "/api/v1/admin/users", nil, staffToken))
	assert.Equal(t, http.StatusForbidden, w.Code,
		"OWASP API5 Secure: staff must be denied /admin/users (403)")
}

func TestAPI5_BFLA_VulnerableMode_Staff_Reaches_Admin_Routes(t *testing.T) {
	// OWASP API5 Vulnerable: AdminOnly middleware bypassed
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)
	app := newTestApp()

	app.userRepo.ListAllFunc = func(_, _ int) ([]models.User, int64, error) { return nil, 0, nil }

	staffToken := helpers.MakeTestToken(uuid.New().String(), "dewi_rahayu", models.RoleStaff)
	w := do(app, helpers.NewJSONRequest("GET", "/api/v1/admin/users", nil, staffToken))
	assert.Equal(t, http.StatusOK, w.Code,
		"OWASP API5 Vulnerable: staff must reach /admin/users (BFLA injection point)")
}

func TestAPI5_BFLA_Nasabah_Cannot_Access_Admin_Routes_SecureMode(t *testing.T) {
	security.SetMode(security.ModeSecure)
	app := newTestApp()
	nasabahToken := helpers.MakeTestToken(uuid.New().String(), "nasabah1", models.RoleNasabah)
	w := do(app, helpers.NewJSONRequest("GET", "/api/v1/admin/users", nil, nasabahToken))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ─────────────────────────────────────────────────────────────────────────────
// OWASP API6 — Unrestricted Access to Sensitive Business Flows
// ─────────────────────────────────────────────────────────────────────────────

func TestAPI6_BusinessFlow_SecureMode_Loan_Limit_Enforced(t *testing.T) {
	// Secure: nasabah cannot have more than N pending loans
	security.SetMode(security.ModeSecure)
	app := newTestApp()

	nasabahID := uuid.New()
	userID := uuid.New()
	nasabah := helpers.MakeTestNasabah(userID)
	nasabah.ID = nasabahID

	app.nasabahRepo.FindByUserIDFunc = func(_ uuid.UUID) (*models.Nasabah, error) { return nasabah, nil }
	// Simulate 3 existing pending loans (limit is 3)
	app.loanRepo.CountPendingByNasabahFunc = func(_ uuid.UUID) (int64, error) { return 3, nil }

	token := helpers.MakeTestToken(userID.String(), "user", models.RoleNasabah)
	w := do(app, helpers.NewJSONRequest("POST", "/api/v1/loans",
		map[string]interface{}{"amount": 5000000, "interest_rate": 12.0, "term_months": 12}, token))

	assert.Equal(t, http.StatusBadRequest, w.Code,
		"OWASP API6 Secure: loan application must be rejected when pending limit is reached")
}

func TestAPI6_BusinessFlow_VulnerableMode_No_Loan_Limit(t *testing.T) {
	// OWASP API6 Vulnerable: no limit on pending applications
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)
	app := newTestApp()

	nasabahID := uuid.New()
	userID := uuid.New()
	nasabah := helpers.MakeTestNasabah(userID)
	nasabah.ID = nasabahID

	app.nasabahRepo.FindByUserIDFunc = func(_ uuid.UUID) (*models.Nasabah, error) { return nasabah, nil }
	// Even with 10 existing pending loans — no check in vulnerable mode
	app.loanRepo.CountPendingByNasabahFunc = func(_ uuid.UUID) (int64, error) { return 10, nil }
	app.loanRepo.CreateFunc = func(_ *models.Loan) error { return nil }

	token := helpers.MakeTestToken(userID.String(), "user", models.RoleNasabah)
	w := do(app, helpers.NewJSONRequest("POST", "/api/v1/loans",
		map[string]interface{}{"amount": 5000000, "interest_rate": 12.0, "term_months": 12}, token))

	assert.Equal(t, http.StatusCreated, w.Code,
		"OWASP API6 Vulnerable: unlimited loan applications must be allowed (no limit enforced)")
}

// ─────────────────────────────────────────────────────────────────────────────
// OWASP API7 — Server-Side Request Forgery (SSRF)
// ─────────────────────────────────────────────────────────────────────────────

func TestAPI7_SSRF_SecureMode_Blocks_Internal_URLs(t *testing.T) {
	// Secure: internal/localhost URLs are blocked
	security.SetMode(security.ModeSecure)
	app := newTestApp()
	adminToken := helpers.MakeTestToken(uuid.New().String(), "budi_santoso", models.RoleAdmin)

	for _, internalURL := range []string{
		"http://localhost:5432",
		"http://127.0.0.1/admin",
		"http://169.254.169.254/latest/meta-data",
		"http://10.0.0.1/internal",
	} {
		w := do(app, helpers.NewJSONRequest("POST", "/api/v1/internal/fetch",
			map[string]interface{}{"url": internalURL}, adminToken))
		// Service returns an error → handler returns 400 (BadRequest), not 403.
		// The important assertion is that it does NOT return 200 (success).
		assert.Equal(t, http.StatusBadRequest, w.Code,
			"OWASP API7 Secure: internal URL %q must be blocked (400 from service error)", internalURL)
	}
}

func TestAPI7_SSRF_VulnerableMode_Allows_Any_URL(t *testing.T) {
	// OWASP API7 Vulnerable: no URL filtering — SSRF is possible
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)
	app := newTestApp()
	adminToken := helpers.MakeTestToken(uuid.New().String(), "budi_santoso", models.RoleAdmin)

	// In vulnerable mode, localhost URL is attempted (will fail TCP but not be blocked)
	w := do(app, helpers.NewJSONRequest("POST", "/api/v1/internal/fetch",
		map[string]interface{}{"url": "http://127.0.0.1:99999"}, adminToken))

	// Should NOT return 403 Forbidden (may return 500/502 from failed TCP, but not 403)
	assert.NotEqual(t, http.StatusForbidden, w.Code,
		"OWASP API7 Vulnerable: internal URLs must NOT be blocked (SSRF injection point)")
}

// ─────────────────────────────────────────────────────────────────────────────
// OWASP API8 — Security Misconfiguration
// ─────────────────────────────────────────────────────────────────────────────

func TestAPI8_SecureHeaders_Present_In_SecureMode_Absent_In_VulnerableMode(t *testing.T) {
	// OWASP API8: In secure mode security headers are set; in vulnerable mode
	// they are intentionally omitted (the vulnerability injection point).

	// Secure: headers must be present
	security.SetMode(security.ModeSecure)
	wSecure := do(newTestApp(), helpers.NewJSONRequest("GET", "/api/system/mode", nil, ""))
	assert.Equal(t, "DENY", wSecure.Header().Get("X-Frame-Options"),
		"API8 Secure: X-Frame-Options must be set")
	assert.Equal(t, "nosniff", wSecure.Header().Get("X-Content-Type-Options"),
		"API8 Secure: X-Content-Type-Options must be set")

	// Vulnerable: security headers are absent (OWASP API8 injection point)
	security.SetMode(security.ModeSandbox)
	wVuln := do(newTestApp(), helpers.NewJSONRequest("GET", "/api/system/mode", nil, ""))
	assert.Empty(t, wVuln.Header().Get("X-Frame-Options"),
		"API8 Vulnerable: X-Frame-Options must be ABSENT (intentional misconfiguration)")
	assert.Empty(t, wVuln.Header().Get("X-Content-Type-Options"),
		"API8 Vulnerable: X-Content-Type-Options must be ABSENT")
	// Vulnerable mode exposes fingerprinting headers instead
	assert.NotEmpty(t, wVuln.Header().Get("X-Security-Mode"),
		"API8 Vulnerable: X-Security-Mode header must be exposed")

	security.SetMode(security.ModeSecure)
}

func TestAPI8_Config_Mode_Endpoint_Returns_Current_Mode(t *testing.T) {
	security.SetMode(security.ModeSecure)
	app := newTestApp()
	w := do(app, helpers.NewJSONRequest("GET", "/config/mode", nil, ""))
	require.Equal(t, http.StatusOK, w.Code)
	body := parseBody(w)
	assert.NotEmpty(t, body["mode"])
}

// ─────────────────────────────────────────────────────────────────────────────
// OWASP API9 — Improper Inventory Management
// ─────────────────────────────────────────────────────────────────────────────

func TestAPI9_DeprecatedV0_SecureMode_Not_Registered(t *testing.T) {
	// OWASP API9 Secure: deprecated v0 routes must return 404 (not registered).
	// We test against a router that only registers v0 routes in vulnerable mode.
	security.SetMode(security.ModeSecure)

	// Build a router that mimics the production behaviour (no v0 in secure mode)
	log := zap.NewNop()
	userRepo := &mocks.MockUserRepository{}
	nasabahRepo := &mocks.MockNasabahRepository{}
	loanRepo := &mocks.MockLoanRepository{}
	txRepo := &mocks.MockTransactionRepository{}
	authSvc := services.NewAuthService(userRepo, helpers.TestJWTSecret, log)
	nasabahSvc := services.NewNasabahService(nasabahRepo, userRepo, log)
	loanSvc := services.NewLoanService(loanRepo, nasabahRepo, log)
	txSvc := services.NewTransactionService(txRepo, loanRepo, nasabahRepo, log)
	loanH := handlers.NewLoanHandler(loanSvc, log)
	adminH := handlers.NewAdminHandler(userRepo, log)
	_ = nasabahSvc
	_ = txSvc

	r := gin.New()
	r.Use(middleware.AuthRequired(authSvc))

	// Register v0 ONLY when in vulnerable mode — same as production app.go
	if security.IsVulnerable() {
		v0 := r.Group("/api/v0")
		v0.GET("/loans", loanH.ListPublic)
		v0.GET("/users", adminH.ListUsersPublic)
		v0.GET("/debug", adminH.Debug)
	}

	adminToken := helpers.MakeTestToken(uuid.New().String(), "budi_santoso", models.RoleAdmin)

	for _, path := range []string{"/api/v0/loans", "/api/v0/users", "/api/v0/debug"} {
		w := helpers.DoRequest(r, helpers.NewJSONRequest("GET", path, nil, adminToken))
		// Either 404 (route not found) or 401 (auth before route check)
		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusUnauthorized,
			"OWASP API9 Secure: deprecated route %q must not be accessible, got %d", path, w.Code)
	}
}

func TestAPI9_DeprecatedV0_VulnerableMode_Exposed(t *testing.T) {
	// OWASP API9 Vulnerable: deprecated v0 routes ARE registered
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)

	log := zap.NewNop()
	userRepo := &mocks.MockUserRepository{
		ListAllFunc: func(_, _ int) ([]models.User, int64, error) { return nil, 0, nil },
	}
	loanRepo := &mocks.MockLoanRepository{
		ListAllFunc: func() ([]models.Loan, error) { return nil, nil },
	}
	authSvc := services.NewAuthService(userRepo, helpers.TestJWTSecret, log)
	loanSvc := services.NewLoanService(loanRepo, &mocks.MockNasabahRepository{},
		log)
	_ = loanSvc
	loanH := handlers.NewLoanHandler(loanSvc, log)
	adminH := handlers.NewAdminHandler(userRepo, log)

	r := gin.New()
	// Register v0 routes — as done in production when vulnerable
	v0 := r.Group("/api/v0")
	v0.GET("/loans", loanH.ListPublic)
	v0.GET("/users", adminH.ListUsersPublic)
	v0.GET("/debug", adminH.Debug)

	_ = authSvc

	for _, path := range []string{"/api/v0/loans", "/api/v0/users", "/api/v0/debug"} {
		w := helpers.DoRequest(r, helpers.NewJSONRequest("GET", path, nil, ""))
		// In vulnerable mode these routes exist and are public (no auth needed)
		assert.NotEqual(t, http.StatusNotFound, w.Code,
			"OWASP API9 Vulnerable: deprecated route %q must be registered and reachable", path)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// OWASP API10 — Unsafe Consumption of APIs
// ─────────────────────────────────────────────────────────────────────────────

func TestAPI10_ResponseFormat_AlwaysJSON(t *testing.T) {
	security.SetMode(security.ModeSecure)
	app := newTestApp()

	for _, path := range []string{"/api/system/mode", "/config/mode"} {
		w := do(app, helpers.NewJSONRequest("GET", path, nil, ""))
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json",
			"all API responses must have application/json Content-Type")
	}
}

func TestAPI10_Error_Response_Is_Structured_JSON(t *testing.T) {
	// API errors must return structured JSON, not HTML or plain text
	security.SetMode(security.ModeSecure)
	app := newTestApp()

	w := do(app, helpers.NewJSONRequest("GET", "/api/v1/auth/me", nil, "invalidtoken"))
	body := parseBody(w)
	assert.NotNil(t, body, "error responses must be valid JSON")
	assert.NotNil(t, body["error"], "error responses must have an 'error' field")
}

func TestAPI10_SecureMode_Consistent_Error_Format(t *testing.T) {
	security.SetMode(security.ModeSecure)
	app := newTestApp()
	user := helpers.MakeTestUser(models.RoleNasabah)
	app.userRepo.FindByUsernameFunc = func(_ string) (*models.User, error) { return user, nil }
	app.userRepo.IncrementLoginAttemptsFunc = func(_ uuid.UUID) error { return nil }

	w := do(app, helpers.NewJSONRequest("POST", "/api/v1/auth/login",
		map[string]interface{}{"username": user.Username, "password": "wrong"}, ""))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	body := parseBody(w)
	_, hasError := body["error"]
	assert.True(t, hasError, "error response must always contain 'error' key")
}

// ─────────────────────────────────────────────────────────────────────────────
// Cross-mode Summary Test
// ─────────────────────────────────────────────────────────────────────────────

func TestOWASP_ModeMatrix(t *testing.T) {
	// Quick matrix: same request → different outcome per mode.
	// Validates that mode switching actually changes system behaviour.
	nasabahID := uuid.New()
	ownerID := uuid.New()
	attackerID := uuid.New()

	type result struct {
		secure     int
		vulnerable int
	}

	// API1 BOLA: attacker accessing owner's nasabah
	t.Run("API1_BOLA", func(t *testing.T) {
		nasabah := helpers.MakeTestNasabah(ownerID)
		var res result

		security.SetMode(security.ModeSecure)
		app := newTestApp()
		app.nasabahRepo.FindByIDFunc = func(_ uuid.UUID) (*models.Nasabah, error) { return nasabah, nil }
		token := helpers.MakeTestToken(attackerID.String(), "attacker", models.RoleNasabah)
		res.secure = helpers.DoRequest(app.router, helpers.NewJSONRequest("GET",
			"/api/v1/nasabah/"+nasabahID.String(), nil, token)).Code

		security.SetMode(security.ModeSandbox)
		app = newTestApp()
		app.nasabahRepo.FindByIDFunc = func(_ uuid.UUID) (*models.Nasabah, error) { return nasabah, nil }
		res.vulnerable = helpers.DoRequest(app.router, helpers.NewJSONRequest("GET",
			"/api/v1/nasabah/"+nasabahID.String(), nil, token)).Code

		security.SetMode(security.ModeSecure)
		assert.Equal(t, http.StatusNotFound, res.secure, "API1: secure blocks IDOR (returns 404 to hide existence)")
		assert.Equal(t, http.StatusOK, res.vulnerable, "API1: vulnerable allows IDOR")
		assert.NotEqual(t, res.secure, res.vulnerable, "API1: modes must differ")
	})

	// API5 BFLA: staff accessing admin route
	t.Run("API5_BFLA", func(t *testing.T) {
		staffToken := helpers.MakeTestToken(uuid.New().String(), "dewi_rahayu", models.RoleStaff)
		var res result

		security.SetMode(security.ModeSecure)
		app := newTestApp()
		res.secure = do(app, helpers.NewJSONRequest("GET", "/api/v1/admin/users", nil, staffToken)).Code

		security.SetMode(security.ModeSandbox)
		app = newTestApp()
		app.userRepo.ListAllFunc = func(_, _ int) ([]models.User, int64, error) { return nil, 0, nil }
		res.vulnerable = do(app, helpers.NewJSONRequest("GET", "/api/v1/admin/users", nil, staffToken)).Code

		security.SetMode(security.ModeSecure)
		assert.Equal(t, http.StatusForbidden, res.secure, "API5: secure enforces RBAC")
		assert.Equal(t, http.StatusOK, res.vulnerable, "API5: vulnerable bypasses RBAC")
	})
}

// ── helpers ───────────────────────────────────────────────────────────────────

var _ repository.NasabahRepository = (*mocks.MockNasabahRepository)(nil)
var _ repository.LoanRepository = (*mocks.MockLoanRepository)(nil)
var _ repository.UserRepository = (*mocks.MockUserRepository)(nil)
var _ repository.TransactionRepository = (*mocks.MockTransactionRepository)(nil)

// compile-time check: mocks satisfy their interfaces
func TestMockInterfaces(t *testing.T) {
	// If this file compiles, all interface assertions above passed.
	t.Log("all mock interfaces compile correctly")
}

// fakeLoanSvc is used by API9 test to avoid wiring up full repo mocks.
type _ = time.Duration // keep time import used
