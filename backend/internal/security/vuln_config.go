package security

import "sync"

// VulnConfig enables fine-grained control over which OWASP Top 10
// vulnerability categories are active while the system runs in
// vulnerable (sandbox) mode.
//
// All categories default to true — identical to the original
// all-or-nothing vulnerable mode, preserving backward compatibility.
//
// INVARIANT: In secure mode this config is ALWAYS ignored.
// IsVulnerableFor / IsSecureFor guarantee this at the call site.
type VulnConfig struct {
	A01_BrokenAccessControl           bool `json:"A01_BrokenAccessControl"`
	A02_CryptographicFailures         bool `json:"A02_CryptographicFailures"`
	A03_Injection                     bool `json:"A03_Injection"`
	A04_InsecureDesign                bool `json:"A04_InsecureDesign"`
	A05_SecurityMisconfiguration      bool `json:"A05_SecurityMisconfiguration"`
	A06_VulnerableComponents          bool `json:"A06_VulnerableComponents"`
	A07_AuthenticationFailures        bool `json:"A07_AuthenticationFailures"`
	A08_SoftwareDataIntegrityFailures bool `json:"A08_SoftwareDataIntegrityFailures"`
	A09_SecurityLoggingFailures       bool `json:"A09_SecurityLoggingFailures"`
	A10_SSRF                          bool `json:"A10_SSRF"`
}

// DefaultVulnConfig returns an all-enabled config — exactly matches the
// behaviour callers got from the original IsVulnerable() checks.
func DefaultVulnConfig() VulnConfig {
	return VulnConfig{
		A01_BrokenAccessControl:           true,
		A02_CryptographicFailures:         true,
		A03_Injection:                     true,
		A04_InsecureDesign:                true,
		A05_SecurityMisconfiguration:      true,
		A06_VulnerableComponents:          true,
		A07_AuthenticationFailures:        true,
		A08_SoftwareDataIntegrityFailures: true,
		A09_SecurityLoggingFailures:       true,
		A10_SSRF:                          true,
	}
}

// ── Category selectors ────────────────────────────────────────────────────────

// CategorySelector extracts one boolean flag from a VulnConfig.
// Pre-defined vars below are used at every injection point so callers
// never write string literals that could silently drift.
type CategorySelector func(VulnConfig) bool

var (
	// CategoryA01 — Broken Object Level Authorization (BOLA / IDOR):
	// ownership checks on nasabah, loan, and transaction resources.
	CategoryA01 CategorySelector = func(c VulnConfig) bool { return c.A01_BrokenAccessControl }

	// CategoryA02 — Cryptographic / Sensitive Data Exposure:
	// controls whether vulnerable response types (including password_hash,
	// NIK, and internal IDs) are returned in API responses.
	CategoryA02 CategorySelector = func(c VulnConfig) bool { return c.A02_CryptographicFailures }

	// CategoryA03 — Injection / BOPLA / Mass Assignment:
	// NIK uniqueness bypass, unchecked field writes.
	CategoryA03 CategorySelector = func(c VulnConfig) bool { return c.A03_Injection }

	// CategoryA04 — Insecure Design / Unrestricted Resource Consumption:
	// pagination enforcement, pending-loan limits, per-window rate gates.
	CategoryA04 CategorySelector = func(c VulnConfig) bool { return c.A04_InsecureDesign }

	// CategoryA05 — Security Misconfiguration:
	// database query logging level, gin debug mode exposure.
	CategoryA05 CategorySelector = func(c VulnConfig) bool { return c.A05_SecurityMisconfiguration }

	// CategoryA06 — Vulnerable / Outdated Components / Business-Flow Abuse:
	// loan amount validation, transaction business-rule enforcement.
	CategoryA06 CategorySelector = func(c VulnConfig) bool { return c.A06_VulnerableComponents }

	// CategoryA07 — Identification and Authentication Failures (BFLA):
	// login plain-text compare, 100-year tokens, account lockout bypass,
	// AdminOnly middleware bypass, loan approve/reject role checks.
	CategoryA07 CategorySelector = func(c VulnConfig) bool { return c.A07_AuthenticationFailures }

	// CategoryA08 — Software and Data Integrity Failures:
	// staff setting loan status to "approved" (bypassing admin-only rule).
	CategoryA08 CategorySelector = func(c VulnConfig) bool { return c.A08_SoftwareDataIntegrityFailures }

	// CategoryA09 — Security Logging and Monitoring Failures:
	// login rate-limit bypass (enables brute-force without detection).
	CategoryA09 CategorySelector = func(c VulnConfig) bool { return c.A09_SecurityLoggingFailures }

	// CategoryA10 — Server-Side Request Forgery (SSRF):
	// URL allowlist bypass in the external-fetch service.
	CategoryA10 CategorySelector = func(c VulnConfig) bool { return c.A10_SSRF }
)

// ── Config store (thread-safe) ────────────────────────────────────────────────

var (
	vulnMu     sync.RWMutex
	vulnConfig = DefaultVulnConfig()
)

// GetVulnConfig returns a copy of the current vulnerability config.
func GetVulnConfig() VulnConfig {
	vulnMu.RLock()
	defer vulnMu.RUnlock()
	return vulnConfig
}

// SetVulnConfig atomically replaces the vulnerability config.
func SetVulnConfig(cfg VulnConfig) {
	vulnMu.Lock()
	defer vulnMu.Unlock()
	vulnConfig = cfg
}

// ResetVulnConfig restores the all-enabled default.
func ResetVulnConfig() { SetVulnConfig(DefaultVulnConfig()) }

// ── Mode-aware per-category helpers ──────────────────────────────────────────

// IsVulnerableFor returns true ONLY when both conditions hold:
//  1. Global mode is "vulnerable" (sandbox).
//  2. The requested OWASP category is enabled in vulnConfig.
//
// Secure mode ALWAYS returns false regardless of config — the invariant
// that "secure mode is never degraded" is enforced here centrally.
func IsVulnerableFor(cat CategorySelector) bool {
	if !IsVulnerable() {
		return false
	}
	vulnMu.RLock()
	defer vulnMu.RUnlock()
	return cat(vulnConfig)
}

// IsSecureFor is the logical complement of IsVulnerableFor.
// Returns true when the system is in secure mode, OR when the
// specific OWASP category has been individually disabled.
func IsSecureFor(cat CategorySelector) bool {
	return !IsVulnerableFor(cat)
}
