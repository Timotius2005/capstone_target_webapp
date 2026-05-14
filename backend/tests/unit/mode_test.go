package unit_test

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"pt-dana-sejahtera/internal/handlers"
	"pt-dana-sejahtera/internal/security"
	"pt-dana-sejahtera/tests/helpers"
)

func setupSystemRouter() *gin.Engine {
	log := zap.NewNop()
	// nil db — system handler is nil-safe for tests (persistence is skipped)
	systemH := handlers.NewSystemHandler(nil, log)
	r := gin.New()
	r.GET("/api/system/mode", systemH.GetMode)
	r.PUT("/api/system/mode", systemH.SetMode)
	return r
}

// ── GET /api/system/mode ──────────────────────────────────────────────────────

func TestGetMode_Returns_CurrentMode_Secure(t *testing.T) {
	security.SetMode(security.ModeSecure)
	w := helpers.DoRequest(setupSystemRouter(), helpers.NewJSONRequest("GET", "/api/system/mode", nil, ""))
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, helpers.ParseBody(w, &body))
	assert.Equal(t, "secure", body["mode"])
}

func TestGetMode_Returns_CurrentMode_Vulnerable(t *testing.T) {
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)
	w := helpers.DoRequest(setupSystemRouter(), helpers.NewJSONRequest("GET", "/api/system/mode", nil, ""))
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, helpers.ParseBody(w, &body))
	// External API uses "vulnerable" not "sandbox"
	assert.Equal(t, "vulnerable", body["mode"])
}

func TestGetMode_NoAuth_Required(t *testing.T) {
	// GET /api/system/mode must be public — no token needed
	security.SetMode(security.ModeSecure)
	w := helpers.DoRequest(setupSystemRouter(), helpers.NewJSONRequest("GET", "/api/system/mode", nil, ""))
	assert.Equal(t, http.StatusOK, w.Code, "GET /api/system/mode must be accessible without token")
}

// ── PUT /api/system/mode ──────────────────────────────────────────────────────

func TestSetMode_Secure_To_Vulnerable(t *testing.T) {
	security.SetMode(security.ModeSecure)
	defer security.SetMode(security.ModeSecure) // restore after test

	w := helpers.DoRequest(setupSystemRouter(), helpers.NewJSONRequest("PUT", "/api/system/mode",
		map[string]interface{}{"mode": "vulnerable"}, ""))

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, helpers.ParseBody(w, &body))
	assert.Equal(t, "vulnerable", body["mode"])

	// Verify in-memory state changed
	assert.True(t, security.IsVulnerable(), "security.IsVulnerable() must be true after switch")
}

func TestSetMode_Vulnerable_To_Secure(t *testing.T) {
	security.SetMode(security.ModeSandbox)
	defer security.SetMode(security.ModeSecure)

	w := helpers.DoRequest(setupSystemRouter(), helpers.NewJSONRequest("PUT", "/api/system/mode",
		map[string]interface{}{"mode": "secure"}, ""))

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, helpers.ParseBody(w, &body))
	assert.Equal(t, "secure", body["mode"])
	assert.True(t, security.IsSecure(), "security.IsSecure() must be true after switch back")
}

func TestSetMode_NoAuth_Required(t *testing.T) {
	// PUT /api/system/mode is intentionally public (pentest lab design)
	security.SetMode(security.ModeSecure)
	defer security.SetMode(security.ModeSecure)
	w := helpers.DoRequest(setupSystemRouter(), helpers.NewJSONRequest("PUT", "/api/system/mode",
		map[string]interface{}{"mode": "vulnerable"}, ""))
	assert.Equal(t, http.StatusOK, w.Code, "PUT /api/system/mode must work without auth token")
}

func TestSetMode_InvalidMode_Returns_400(t *testing.T) {
	security.SetMode(security.ModeSecure)
	w := helpers.DoRequest(setupSystemRouter(), helpers.NewJSONRequest("PUT", "/api/system/mode",
		map[string]interface{}{"mode": "hacked"}, ""))
	assert.Equal(t, http.StatusBadRequest, w.Code, "unknown mode values must be rejected")
}

func TestSetMode_EmptyBody_Returns_400(t *testing.T) {
	security.SetMode(security.ModeSecure)
	w := helpers.DoRequest(setupSystemRouter(), helpers.NewJSONRequest("PUT", "/api/system/mode", nil, ""))
	assert.Equal(t, http.StatusBadRequest, w.Code, "missing mode field must return 400")
}

func TestSetMode_AcceptsSandboxAlias(t *testing.T) {
	// "sandbox" is a legacy alias for "vulnerable"
	security.SetMode(security.ModeSecure)
	defer security.SetMode(security.ModeSecure)
	w := helpers.DoRequest(setupSystemRouter(), helpers.NewJSONRequest("PUT", "/api/system/mode",
		map[string]interface{}{"mode": "sandbox"}, ""))
	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, security.IsVulnerable())
}

// ── In-memory mode state ──────────────────────────────────────────────────────

func TestSecurityPackage_SetMode_IsIdempotent(t *testing.T) {
	security.SetMode(security.ModeSecure)
	assert.True(t, security.IsSecure())
	assert.False(t, security.IsVulnerable())

	security.SetMode(security.ModeSandbox)
	assert.False(t, security.IsSecure())
	assert.True(t, security.IsVulnerable())

	security.SetMode(security.ModeSecure) // restore
	assert.True(t, security.IsSecure())
}

func TestSecurityPackage_GetMode_Returns_Correct_String(t *testing.T) {
	security.SetMode(security.ModeSecure)
	assert.Equal(t, "secure", security.GetMode())

	security.SetMode(security.ModeSandbox)
	assert.Equal(t, "sandbox", security.GetMode())

	security.SetMode(security.ModeSecure) // restore
}

func TestSecurityPackage_ModeChange_NoRestart_Required(t *testing.T) {
	// Mode switching happens in-memory — no server restart needed.
	// Rapid switches must all take effect.
	for i := 0; i < 10; i++ {
		security.SetMode(security.ModeSandbox)
		assert.True(t, security.IsVulnerable(), "iteration %d: expected vulnerable", i)
		security.SetMode(security.ModeSecure)
		assert.True(t, security.IsSecure(), "iteration %d: expected secure", i)
	}
}
