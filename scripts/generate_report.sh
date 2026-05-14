#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────────────────────
# generate_report.sh — Consolidate test artefacts into reports/test-report.md
# ──────────────────────────────────────────────────────────────────────────────
set -euo pipefail

REPORTS_DIR="$(dirname "$0")/../reports"
OUTPUT="$REPORTS_DIR/test-report.md"
NOW=$(date '+%Y-%m-%d %H:%M:%S')

mkdir -p "$REPORTS_DIR"

cat > "$OUTPUT" << EOF
# PT. Dana Sejahtera — Test Report

**Generated:** $NOW
**Project:** PT. Dana Sejahtera Fintech Security Evaluation Platform
**Repository:** capstone_target_webapp

---

## Summary

| Suite | Status | Details |
|-------|--------|---------|
EOF

# ── Backend unit tests ────────────────────────────────────────────────────────

if [ -f "$REPORTS_DIR/backend-unit.log" ]; then
  PASS_COUNT=$(grep -c "^--- PASS:" "$REPORTS_DIR/backend-unit.log" 2>/dev/null || echo 0)
  FAIL_COUNT=$(grep -c "^--- FAIL:" "$REPORTS_DIR/backend-unit.log" 2>/dev/null || echo 0)
  STATUS="✅ PASSED"
  [ "$FAIL_COUNT" -gt 0 ] && STATUS="❌ FAILED"
  echo "| Backend Unit Tests | $STATUS | Pass: $PASS_COUNT  Fail: $FAIL_COUNT |" >> "$OUTPUT"
else
  echo "| Backend Unit Tests | ⚠ Not run | — |" >> "$OUTPUT"
fi

# ── Backend OWASP tests ───────────────────────────────────────────────────────

if [ -f "$REPORTS_DIR/backend-owasp.log" ]; then
  PASS_COUNT=$(grep -c "^--- PASS:" "$REPORTS_DIR/backend-owasp.log" 2>/dev/null || echo 0)
  FAIL_COUNT=$(grep -c "^--- FAIL:" "$REPORTS_DIR/backend-owasp.log" 2>/dev/null || echo 0)
  STATUS="✅ PASSED"
  [ "$FAIL_COUNT" -gt 0 ] && STATUS="❌ FAILED"
  echo "| Backend OWASP Tests | $STATUS | Pass: $PASS_COUNT  Fail: $FAIL_COUNT |" >> "$OUTPUT"
else
  echo "| Backend OWASP Tests | ⚠ Not run | — |" >> "$OUTPUT"
fi

# ── Frontend unit tests ───────────────────────────────────────────────────────

if [ -f "$REPORTS_DIR/frontend-unit.log" ]; then
  PASS_COUNT=$(grep -oP "(?<=Tests:)\s+\K[0-9]+ passed" "$REPORTS_DIR/frontend-unit.log" 2>/dev/null | head -1 || echo "?")
  STATUS="✅ PASSED"
  grep -q "FAIL" "$REPORTS_DIR/frontend-unit.log" 2>/dev/null && STATUS="❌ FAILED"
  echo "| Frontend Jest Tests | $STATUS | $PASS_COUNT |" >> "$OUTPUT"
else
  echo "| Frontend Jest Tests | ⚠ Not run | — |" >> "$OUTPUT"
fi

# ── Security attack simulator ─────────────────────────────────────────────────

if [ -f "$REPORTS_DIR/security-validation.json" ]; then
  SECURE_PASS=$(python3 -c "
import json, sys
with open('$REPORTS_DIR/security-validation.json') as f:
    data = json.load(f)
print(data['summary']['secure_passed'])
" 2>/dev/null || echo "?")
  SECURE_FAIL=$(python3 -c "
import json, sys
with open('$REPORTS_DIR/security-validation.json') as f:
    data = json.load(f)
print(data['summary']['secure_failed'])
" 2>/dev/null || echo "?")
  STATUS="✅ PASSED"
  [ "$SECURE_FAIL" != "0" ] && [ "$SECURE_FAIL" != "?" ] && STATUS="❌ FAILED"
  echo "| Security Attack Simulator | $STATUS | Secure pass: $SECURE_PASS  Fail: $SECURE_FAIL |" >> "$OUTPUT"
else
  echo "| Security Attack Simulator | ⚠ Not run | — |" >> "$OUTPUT"
fi

# ── OWASP Security Matrix ─────────────────────────────────────────────────────

cat >> "$OUTPUT" << 'EOF'

---

## OWASP API Security Top 10 Matrix

| # | Category | Secure Mode Expected | Vulnerable Mode Expected |
|---|----------|---------------------|-------------------------|
| API1 | Broken Object Level Auth | ✅ Blocked (403) | ✅ Allowed (200) |
| API2 | Broken Authentication | ✅ bcrypt + lockout | ✅ Plaintext + no lockout |
| API3 | Broken Object Property Level Auth | ✅ NIK masked, no password_hash | ✅ Full NIK + password_hash exposed |
| API4 | Unrestricted Resource Consumption | ✅ Paginated + rate-limited | ✅ Full table dump, no limit |
| API5 | Broken Function Level Auth | ✅ RBAC enforced (403) | ✅ Admin bypass (200) |
| API6 | Unrestricted Business Flows | ✅ Loan limit enforced | ✅ Unlimited applications |
| API7 | Server-Side Request Forgery | ✅ Internal URLs blocked | ✅ Any URL reachable |
| API8 | Security Misconfiguration | ✅ Headers present | ⚠ Same (headers always set) |
| API9 | Improper Inventory Management | ✅ /api/v0 → 404 | ✅ /api/v0 routes exposed |
| API10 | Unsafe API Consumption | ✅ Structured JSON errors | ✅ Same (always JSON) |

EOF

# ── Test files listing ────────────────────────────────────────────────────────

cat >> "$OUTPUT" << 'EOF'

---

## Test File Structure

```
backend/tests/
├── helpers/helpers.go          # JWT creation, request helpers, test factories
├── mocks/mocks.go              # Mock repository implementations
├── unit/
│   ├── auth_test.go            # Auth handler + service (login, register, token)
│   ├── mode_test.go            # Mode switching — GET/PUT /api/system/mode
│   └── middleware_test.go      # AuthRequired, AdminOnly, RoleCheck
└── owasp/
    └── owasp_test.go           # All 10 OWASP categories, both modes

frontend/
├── __tests__/
│   ├── ModeBadge.test.tsx      # Badge rendering + colour per mode
│   ├── GlobalModeSwitcher.test.tsx # Toggle interaction, toasts, accessibility
│   └── ModeContext.test.tsx    # Context fetch, switchMode, cookie sync
└── e2e/
    └── mode-toggle.spec.ts     # Playwright: full mode-switch flow end-to-end

tests/security/
└── attack_simulator.py         # Python OWASP attack simulation script
```

EOF

# ── Coverage note ─────────────────────────────────────────────────────────────

cat >> "$OUTPUT" << 'EOF'

---

## Coverage Reports

| Report | Location |
|--------|----------|
| Backend unit coverage (HTML) | `reports/backend-unit.coverage` |
| Backend OWASP coverage (HTML) | `reports/backend-owasp.coverage` |
| Frontend Jest coverage (HTML) | `reports/frontend-coverage/lcov-report/` |
| Playwright report (HTML) | `reports/playwright-report/` |
| Security JSON matrix | `reports/security-validation.json` |

> Run `go tool cover -html=reports/backend-unit.coverage -o reports/backend-coverage.html`
> to view backend coverage as HTML.

---

*Generated by `make report` — PT. Dana Sejahtera Security Testing Platform*
EOF

echo "Report written to: $OUTPUT"
