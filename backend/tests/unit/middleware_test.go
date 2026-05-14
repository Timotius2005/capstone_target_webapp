package unit_test

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"pt-dana-sejahtera/internal/middleware"
	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/security"
	"pt-dana-sejahtera/internal/services"
	"pt-dana-sejahtera/tests/helpers"
	"pt-dana-sejahtera/tests/mocks"
)

// protectedRouter wraps a single 200-OK handler behind auth and role middleware.
func protectedRouter(authSvc services.AuthService, extra ...gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	chain := append([]gin.HandlerFunc{middleware.AuthRequired(authSvc)}, extra...)
	chain = append(chain, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/protected", chain...)
	return r
}

func makeAuthSvc(userRepo *mocks.MockUserRepository) services.AuthService {
	return services.NewAuthService(userRepo, helpers.TestJWTSecret, zap.NewNop())
}

// ── AuthRequired ──────────────────────────────────────────────────────────────

func TestAuthRequired_MissingHeader_Returns_401(t *testing.T) {
	security.SetMode(security.ModeSecure)
	r := protectedRouter(makeAuthSvc(&mocks.MockUserRepository{}))
	w := helpers.DoRequest(r, helpers.NewJSONRequest("GET", "/protected", nil, ""))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_MalformedHeader_Returns_401(t *testing.T) {
	security.SetMode(security.ModeSecure)
	r := protectedRouter(makeAuthSvc(&mocks.MockUserRepository{}))
	req := helpers.NewJSONRequest("GET", "/protected", nil, "")
	req.Header.Set("Authorization", "NotBearer token123")
	w := helpers.DoRequest(r, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_ValidToken_Passes(t *testing.T) {
	security.SetMode(security.ModeSecure)
	user := helpers.MakeTestUser(models.RoleNasabah)
	token := helpers.MakeTestToken(user.ID.String(), user.Username, user.Role)
	r := protectedRouter(makeAuthSvc(&mocks.MockUserRepository{}))
	w := helpers.DoRequest(r, helpers.NewJSONRequest("GET", "/protected", nil, token))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthRequired_ExpiredToken_Returns_401(t *testing.T) {
	security.SetMode(security.ModeSecure)
	user := helpers.MakeTestUser(models.RoleNasabah)
	expired := helpers.MakeExpiredToken(user.ID.String(), user.Username, user.Role)
	r := protectedRouter(makeAuthSvc(&mocks.MockUserRepository{}))
	w := helpers.DoRequest(r, helpers.NewJSONRequest("GET", "/protected", nil, expired))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_BothModes_AlwaysEnforced(t *testing.T) {
	// Auth is NEVER disabled — both modes require a valid JWT.
	for _, mode := range []security.ModeValue{security.ModeSecure, security.ModeSandbox} {
		security.SetMode(mode)
		r := protectedRouter(makeAuthSvc(&mocks.MockUserRepository{}))
		w := helpers.DoRequest(r, helpers.NewJSONRequest("GET", "/protected", nil, ""))
		assert.Equal(t, http.StatusUnauthorized, w.Code,
			"AuthRequired must reject empty token in mode=%s", mode)
	}
	security.SetMode(security.ModeSecure) // restore
}

// ── AdminOnly ─────────────────────────────────────────────────────────────────

func TestAdminOnly_SecureMode_NonAdmin_Returns_403(t *testing.T) {
	// OWASP API5 Secure: non-admin must be denied
	security.SetMode(security.ModeSecure)
	user := helpers.MakeTestUser(models.RoleStaff)
	token := helpers.MakeTestToken(user.ID.String(), user.Username, user.Role)

	r := protectedRouter(makeAuthSvc(&mocks.MockUserRepository{}), middleware.AdminOnly())
	w := helpers.DoRequest(r, helpers.NewJSONRequest("GET", "/protected", nil, token))
	assert.Equal(t, http.StatusForbidden, w.Code,
		"OWASP API5: staff must be denied admin routes in secure mode")
}

func TestAdminOnly_SecureMode_Admin_Passes(t *testing.T) {
	security.SetMode(security.ModeSecure)
	user := helpers.MakeTestUser(models.RoleAdmin)
	token := helpers.MakeTestToken(user.ID.String(), user.Username, user.Role)

	r := protectedRouter(makeAuthSvc(&mocks.MockUserRepository{}), middleware.AdminOnly())
	w := helpers.DoRequest(r, helpers.NewJSONRequest("GET", "/protected", nil, token))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminOnly_VulnerableMode_NonAdmin_Bypasses(t *testing.T) {
	// OWASP API5 Vulnerable: AdminOnly is bypassed — any authenticated user passes
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)

	user := helpers.MakeTestUser(models.RoleNasabah) // lowest privilege
	token := helpers.MakeTestToken(user.ID.String(), user.Username, user.Role)

	r := protectedRouter(makeAuthSvc(&mocks.MockUserRepository{}), middleware.AdminOnly())
	w := helpers.DoRequest(r, helpers.NewJSONRequest("GET", "/protected", nil, token))
	assert.Equal(t, http.StatusOK, w.Code,
		"OWASP API5: AdminOnly must be bypassed in vulnerable mode (BFLA injection point)")
}

func TestAdminOnly_Mode_Difference_Is_Observable(t *testing.T) {
	user := helpers.MakeTestUser(models.RoleStaff)
	token := helpers.MakeTestToken(user.ID.String(), user.Username, user.Role)

	// Secure: 403
	security.SetMode(security.ModeSecure)
	r := protectedRouter(makeAuthSvc(&mocks.MockUserRepository{}), middleware.AdminOnly())
	wSecure := helpers.DoRequest(r, helpers.NewJSONRequest("GET", "/protected", nil, token))

	// Vulnerable: 200
	security.SetMode(security.ModeSandbox)
	r = protectedRouter(makeAuthSvc(&mocks.MockUserRepository{}), middleware.AdminOnly())
	wVulnerable := helpers.DoRequest(r, helpers.NewJSONRequest("GET", "/protected", nil, token))

	security.SetMode(security.ModeSecure)

	assert.Equal(t, http.StatusForbidden, wSecure.Code,
		"secure mode: staff blocked from admin route")
	assert.Equal(t, http.StatusOK, wVulnerable.Code,
		"vulnerable mode: staff reaches admin route (OWASP API5 BFLA)")
	assert.NotEqual(t, wSecure.Code, wVulnerable.Code,
		"modes must produce different status codes for BFLA test")
}

// ── RoleCheck ─────────────────────────────────────────────────────────────────

func TestRoleCheck_Always_Enforced_In_Both_Modes(t *testing.T) {
	// RoleCheck is NEVER bypassed — unlike AdminOnly it always checks role.
	for _, mode := range []security.ModeValue{security.ModeSecure, security.ModeSandbox} {
		security.SetMode(mode)
		user := helpers.MakeTestUser(models.RoleNasabah)
		token := helpers.MakeTestToken(user.ID.String(), user.Username, user.Role)

		r := protectedRouter(makeAuthSvc(&mocks.MockUserRepository{}),
			middleware.RoleCheck(models.RoleAdmin)) // requires admin
		w := helpers.DoRequest(r, helpers.NewJSONRequest("GET", "/protected", nil, token))
		assert.Equal(t, http.StatusForbidden, w.Code,
			"RoleCheck must always deny wrong role (mode=%s)", mode)
	}
	security.SetMode(security.ModeSecure)
}

func TestRoleCheck_Allows_Correct_Role(t *testing.T) {
	security.SetMode(security.ModeSecure)
	user := helpers.MakeTestUser(models.RoleAdmin)
	token := helpers.MakeTestToken(user.ID.String(), user.Username, user.Role)

	r := protectedRouter(makeAuthSvc(&mocks.MockUserRepository{}),
		middleware.RoleCheck(models.RoleAdmin, models.RoleStaff))
	w := helpers.DoRequest(r, helpers.NewJSONRequest("GET", "/protected", nil, token))
	assert.Equal(t, http.StatusOK, w.Code)
}

// ── Context injection ─────────────────────────────────────────────────────────

func TestAuthRequired_Injects_Claims_Into_Context(t *testing.T) {
	security.SetMode(security.ModeSecure)
	user := helpers.MakeTestUser(models.RoleAdmin)
	token := helpers.MakeTestToken(user.ID.String(), user.Username, user.Role)

	var capturedRole string
	r := gin.New()
	r.GET("/check", middleware.AuthRequired(makeAuthSvc(&mocks.MockUserRepository{})),
		func(c *gin.Context) {
			capturedRole = c.GetString(middleware.ContextRole)
			c.JSON(http.StatusOK, gin.H{})
		})

	helpers.DoRequest(r, helpers.NewJSONRequest("GET", "/check", nil, token))
	assert.Equal(t, models.RoleAdmin, capturedRole,
		"AuthRequired must inject role into Gin context")
}
