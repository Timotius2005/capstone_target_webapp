package unit_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/security"
	"pt-dana-sejahtera/tests/helpers"
	"pt-dana-sejahtera/tests/mocks"
)

// resetConfig is a test helper that restores defaults after each sub-test.
func resetConfig(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		security.ResetVulnConfig()
		security.SetMode(security.ModeSecure)
	})
}

// ── VulnConfig struct & defaults ──────────────────────────────────────────────

func TestVulnConfig_DefaultsAllTrue(t *testing.T) {
	cfg := security.DefaultVulnConfig()

	assert.True(t, cfg.A01_BrokenAccessControl, "A01 must default to true")
	assert.True(t, cfg.A02_CryptographicFailures, "A02 must default to true")
	assert.True(t, cfg.A03_Injection, "A03 must default to true")
	assert.True(t, cfg.A04_InsecureDesign, "A04 must default to true")
	assert.True(t, cfg.A05_SecurityMisconfiguration, "A05 must default to true")
	assert.True(t, cfg.A06_VulnerableComponents, "A06 must default to true")
	assert.True(t, cfg.A07_AuthenticationFailures, "A07 must default to true")
	assert.True(t, cfg.A08_SoftwareDataIntegrityFailures, "A08 must default to true")
	assert.True(t, cfg.A09_SecurityLoggingFailures, "A09 must default to true")
	assert.True(t, cfg.A10_SSRF, "A10 must default to true")
}

func TestVulnConfig_SetAndGet_RoundTrip(t *testing.T) {
	resetConfig(t)
	custom := security.DefaultVulnConfig()
	custom.A01_BrokenAccessControl = false
	custom.A10_SSRF = false

	security.SetVulnConfig(custom)
	got := security.GetVulnConfig()

	assert.False(t, got.A01_BrokenAccessControl)
	assert.True(t, got.A02_CryptographicFailures)
	assert.False(t, got.A10_SSRF)
}

func TestVulnConfig_ResetRestoresDefaults(t *testing.T) {
	resetConfig(t)
	cfg := security.DefaultVulnConfig()
	cfg.A07_AuthenticationFailures = false
	security.SetVulnConfig(cfg)

	security.ResetVulnConfig()

	assert.True(t, security.GetVulnConfig().A07_AuthenticationFailures,
		"ResetVulnConfig must restore A07 to true")
}

// ── IsVulnerableFor / IsSecureFor in secure mode ──────────────────────────────

func TestIsVulnerableFor_SecureMode_AlwaysFalse(t *testing.T) {
	resetConfig(t)
	security.SetMode(security.ModeSecure)

	// Even with all categories enabled, secure mode must never return vulnerable.
	security.SetVulnConfig(security.DefaultVulnConfig())

	assert.False(t, security.IsVulnerableFor(security.CategoryA01),
		"secure mode: A01 must not be vulnerable regardless of config")
	assert.False(t, security.IsVulnerableFor(security.CategoryA07),
		"secure mode: A07 must not be vulnerable regardless of config")
	assert.False(t, security.IsVulnerableFor(security.CategoryA10),
		"secure mode: A10 must not be vulnerable regardless of config")
}

func TestIsSecureFor_SecureMode_AlwaysTrue(t *testing.T) {
	resetConfig(t)
	security.SetMode(security.ModeSecure)

	assert.True(t, security.IsSecureFor(security.CategoryA01),
		"secure mode: IsSecureFor(A01) must always be true")
	assert.True(t, security.IsSecureFor(security.CategoryA09),
		"secure mode: IsSecureFor(A09) must always be true")
}

// ── IsVulnerableFor in vulnerable mode ───────────────────────────────────────

func TestIsVulnerableFor_VulnerableMode_DefaultConfig_AllTrue(t *testing.T) {
	resetConfig(t)
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)

	security.ResetVulnConfig() // all true

	assert.True(t, security.IsVulnerableFor(security.CategoryA01))
	assert.True(t, security.IsVulnerableFor(security.CategoryA07))
	assert.True(t, security.IsVulnerableFor(security.CategoryA10))
}

func TestIsVulnerableFor_VulnerableMode_DisabledCategory_ReturnsFalse(t *testing.T) {
	resetConfig(t)
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)

	cfg := security.DefaultVulnConfig()
	cfg.A07_AuthenticationFailures = false
	security.SetVulnConfig(cfg)

	assert.False(t, security.IsVulnerableFor(security.CategoryA07),
		"disabled A07 must behave as secure in vulnerable mode")
	assert.True(t, security.IsVulnerableFor(security.CategoryA01),
		"enabled A01 must remain vulnerable when only A07 is disabled")
}

// ── Mixed state: one category OFF does not affect others ─────────────────────

func TestVulnConfig_MixedState_OnlyTargetedCategorySecure(t *testing.T) {
	resetConfig(t)
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)

	cfg := security.DefaultVulnConfig()
	cfg.A01_BrokenAccessControl = false // only A01 disabled

	security.SetVulnConfig(cfg)

	// A01 should now behave as secure.
	assert.False(t, security.IsVulnerableFor(security.CategoryA01),
		"A01 disabled → must not be vulnerable")
	assert.True(t, security.IsSecureFor(security.CategoryA01),
		"A01 disabled → IsSecureFor must be true")

	// All other categories must remain vulnerable.
	assert.True(t, security.IsVulnerableFor(security.CategoryA02), "A02 unaffected")
	assert.True(t, security.IsVulnerableFor(security.CategoryA07), "A07 unaffected")
	assert.True(t, security.IsVulnerableFor(security.CategoryA09), "A09 unaffected")
	assert.True(t, security.IsVulnerableFor(security.CategoryA10), "A10 unaffected")
}

// ── Auth service: A07 toggle controls login behaviour ────────────────────────

func TestAuthLogin_A07Disabled_UsesBcryptAndGenericError(t *testing.T) {
	// When mode=vulnerable but A07 is OFF, the login must behave as secure:
	// bcrypt comparison, generic error, proper response format.
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)

	cfg := security.DefaultVulnConfig()
	cfg.A07_AuthenticationFailures = false
	security.SetVulnConfig(cfg)
	defer security.ResetVulnConfig()

	user := helpers.MakeTestUser(models.RoleNasabah)
	// PasswordHash is a bcrypt hash of "TestPass123!" — NOT plain text.
	incrementCalled := false
	userRepo := &mocks.MockUserRepository{
		FindByUsernameFunc:         func(_ string) (*models.User, error) { return user, nil },
		IncrementLoginAttemptsFunc: func(_ uuid.UUID) error { incrementCalled = true; return nil },
	}

	w := helpers.DoRequest(setupAuthRouter(userRepo), helpers.NewJSONRequest(
		"POST", "/auth/login",
		map[string]interface{}{"username": user.Username, "password": "wrongpassword"},
		"",
	))

	require.Equal(t, http.StatusUnauthorized, w.Code,
		"A07 disabled in vulnerable mode: wrong password must be rejected")

	var body map[string]interface{}
	require.NoError(t, helpers.ParseBody(w, &body))
	assert.Equal(t, "invalid credentials", body["error"],
		"A07 disabled: error must be generic (not reveal username)")
	assert.True(t, incrementCalled,
		"A07 disabled: failed attempt counter must be incremented")
}

func TestAuthLogin_A07Enabled_UsesPlaintextAndVerboseError(t *testing.T) {
	// When mode=vulnerable and A07 is ON, the login must use vulnerable behaviour.
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)
	security.ResetVulnConfig() // all ON
	defer security.ResetVulnConfig()

	user := helpers.MakeTestUser(models.RoleNasabah)
	user.PasswordHash = "correctplaintext" // plain text stored in hash field

	userRepo := &mocks.MockUserRepository{
		FindByUsernameFunc: func(_ string) (*models.User, error) { return user, nil },
	}

	w := helpers.DoRequest(setupAuthRouter(userRepo), helpers.NewJSONRequest(
		"POST", "/auth/login",
		map[string]interface{}{"username": user.Username, "password": "wrongpassword"},
		"",
	))

	require.Equal(t, http.StatusUnauthorized, w.Code)
	var body map[string]interface{}
	require.NoError(t, helpers.ParseBody(w, &body))
	assert.Contains(t, body["error"].(string), user.Username,
		"A07 enabled: verbose error must include username")
}

// ── SetVulnConfig rejected in secure mode (via system handler) ───────────────

func TestSystemHandler_SetVulnConfig_RequiresVulnerableMode(t *testing.T) {
	security.SetMode(security.ModeSecure)

	r := setupSystemRouter()
	cfg := security.DefaultVulnConfig()
	w := helpers.DoRequest(r, helpers.NewJSONRequest(
		"PUT", "/api/system/vuln-config", cfg, "",
	))

	assert.Equal(t, http.StatusBadRequest, w.Code,
		"SetVulnConfig must return 400 when mode=secure")
}

func TestSystemHandler_SetVulnConfig_AcceptedInVulnerableMode(t *testing.T) {
	security.SetMode(security.ModeSandbox)
	defer func() {
		security.SetMode(security.ModeSecure)
		security.ResetVulnConfig()
	}()

	r := setupSystemRouter()
	cfg := security.DefaultVulnConfig()
	cfg.A01_BrokenAccessControl = false

	w := helpers.DoRequest(r, helpers.NewJSONRequest(
		"PUT", "/api/system/vuln-config", cfg, "",
	))

	assert.Equal(t, http.StatusOK, w.Code, "SetVulnConfig must succeed in vulnerable mode")
	assert.False(t, security.GetVulnConfig().A01_BrokenAccessControl,
		"A01 must be persisted as false after PUT")
}

func TestSystemHandler_GetVulnConfig_AlwaysAccessible(t *testing.T) {
	for _, mode := range []security.ModeValue{security.ModeSecure, security.ModeSandbox} {
		security.SetMode(mode)
		r := setupSystemRouter()
		w := helpers.DoRequest(r, helpers.NewJSONRequest("GET", "/api/system/vuln-config", nil, ""))
		assert.Equal(t, http.StatusOK, w.Code,
			"GET /api/system/vuln-config must be accessible in mode=%s", mode)
	}
	security.SetMode(security.ModeSecure)
}
