# PT. Dana Sejahtera — Full-Stack Test Report

**Generated:** 2026-05-15  
**Project:** PT. Dana Sejahtera Fintech Security Evaluation Platform  
**Repository:** capstone_target_webapp

---

## Credential Update Summary

| Role  | Old Username | Old Password | New Username   | New Password        |
|-------|--------------|--------------|----------------|---------------------|
| admin | `admin`      | `Admin@123`  | `budi_santoso` | `Admin@Dana2025!`   |
| staff | `staff01`    | `Admin@123`  | `dewi_rahayu`  | `Staff@Dana2025!`   |

All passwords hashed with **bcrypt, cost 10** — identical algorithm to production.

---

## 1. Backend Unit Tests

**Runner:** `go test ./tests/... -coverpkg=./internal/... -count=1`  
**Result:** ✅ **65 / 65 PASSED — 0 FAILED**

| Package          | Tests | Passed | Failed | Duration |
|------------------|-------|--------|--------|----------|
| `tests/unit`     | 34    | 34     | 0      | ~0.97 s  |
| `tests/owasp`    | 31    | 31     | 0      | ~0.20 s  |
| **Total**        | **65**| **65** | **0**  | ~1.17 s  |

### Coverage (against `./internal/...`)

| Suite        | Coverage |
|--------------|----------|
| OWASP tests  | 25.8 %   |
| Unit tests   | 13.3 %   |
| **Combined** | **31.0 %**|

HTML report: `reports/coverage-backend.html`

### Test categories covered

| Category | Tests | Result |
|----------|-------|--------|
| Registration (secure/duplicate/weak-pw) | 3 | ✅ |
| Login — secure mode (valid, wrong pw, locked) | 3 | ✅ |
| Login — vulnerable mode (plain-text, verbose err, no lockout) | 3 | ✅ |
| JWT validation (valid / expired / tampered / missing) | 4 | ✅ |
| AuthRequired middleware (missing/malformed/valid/expired/both-modes) | 5 | ✅ |
| AdminOnly middleware (secure deny, secure pass, vulnerable bypass) | 4 | ✅ |
| RoleCheck middleware (always enforced, correct role) | 2 | ✅ |
| Context injection | 1 | ✅ |
| System mode GET/PUT/aliases/validation | 8 | ✅ |
| In-memory mode state (idempotent, string, rapid toggle) | 3 | ✅ |
| OWASP API1 BOLA/IDOR (secure block, vulnerable allow, admin exempt) | 3 | ✅ |
| OWASP API2 Broken Auth (bcrypt, timing, plaintext) | 3 | ✅ |
| OWASP API3 BOPLA (response exposure) | 2 | ✅ |
| OWASP API4 Resource Consumption (rate-limit toggle) | 2 | ✅ |
| OWASP API5 BFLA (staff→admin denied/bypassed, nasabah denied) | 3 | ✅ |
| OWASP API6 Sensitive Business Flows | 2 | ✅ |
| OWASP API7 SSRF (blocked/allowed) | 2 | ✅ |
| OWASP API8 Security Misconfiguration | 2 | ✅ |
| OWASP API9 Inventory (v0 deprecated routes) | 2 | ✅ |
| OWASP API10 Unsafe Consumption | 1 | ✅ |
| Cross-OWASP integration sweep | 1 | ✅ |

---

## 2. Frontend Unit Tests (Jest / React Testing Library)

**Runner:** `npm test -- --forceExit`  
**Result:** ✅ **31 / 31 PASSED — 0 FAILED**

| Suite                       | Tests | Passed | Failed | Duration |
|-----------------------------|-------|--------|--------|----------|
| `GlobalModeSwitcher.test`   | 10    | 10     | 0      |          |
| `ModeBadge.test`            | 11    | 11     | 0      |          |
| `ModeContext.test`          | 10    | 10     | 0      |          |
| **Total**                   | **31**| **31** | **0**  | ~3.0 s   |

### Coverage

| File                   | Stmts  | Branch | Funcs  | Lines  |
|------------------------|--------|--------|--------|--------|
| GlobalModeSwitcher.tsx | 90.0 % | 93.9 % | 66.7 % | 100 %  |
| ModeBadge.tsx          | 100 %  | 100 %  | 100 %  | 100 %  |
| ModeContext.tsx        | 61.8 % | 27.6 % | 66.7 % | 63.5 % |
| **All files**          | 12.0 % | 10.1 % | 6.4 %  | 12.5 % |

---

## 3. E2E Tests (Playwright — Chromium)

**Runner:** `npm run test:e2e`  
**Config:** `workers: 1`, `fullyParallel: false`, `globalSetup` seeds users and captures auth state  
**Result:** ✅ **18 / 18 PASSED — 0 FAILED**

| Suite                                       | Tests | Passed | Failed | Duration |
|---------------------------------------------|-------|--------|--------|----------|
| GlobalModeSwitcher banner                   | 5     | 5      | 0      | ~3.8 s   |
| Mode toggle — login page (no auth required) | 4     | 4      | 0      | ~2.1 s   |
| Mode toggle — dashboard (authenticated)     | 4     | 4      | 0      | ~5.9 s   |
| GlobalModeSwitcher — accessibility          | 3     | 3      | 0      | ~1.3 s   |
| ModeBadge in Navbar (authenticated)         | 2     | 2      | 0      | ~2.1 s   |
| **Total**                                   | **18**| **18** | **0**  | ~17.8 s  |

### Auth strategy

- `globalSetup` registers `budi_santoso` (admin) and `dewi_rahayu` (staff) via `/api/v1/auth/register`  
- Performs **one** browser login per role and persists `storageState` to `e2e/.auth/`  
- Authenticated test suites load stored state — no repeated UI logins → stays within backend rate limit (5 req / 60 s)  
- Admin (`budi_santoso`) login redirects to `/dashboard` ✅  
- Staff (`dewi_rahayu`) state captured and available for future staff-specific suites ✅  

---

## 4. Grand Total

| Suite              | Tests | Passed | Failed |
|--------------------|-------|--------|--------|
| Backend unit       | 65    | 65     | 0      |
| Frontend unit      | 31    | 31     | 0      |
| E2E (Playwright)   | 18    | 18     | 0      |
| **Grand Total**    | **114**| **114**| **0** |

---

## 5. Files Added / Changed

| File | Change |
|------|--------|
| `backend/migrations/000003_test_users.up.sql` | **New** — seeds `budi_santoso` + `dewi_rahayu` with bcrypt hashes |
| `backend/migrations/000003_test_users.down.sql` | **New** — rollback |
| `backend/pkg/seed/seed.go` | **New** — programmatic Go seeder (bcrypt cost 10 at runtime) |
| `backend/tests/owasp/owasp_test.go` | JWT claim usernames: `admin`→`budi_santoso`, `staff01`/`staff`→`dewi_rahayu` |
| `frontend/e2e/global-setup.ts` | New default credentials; seeds both roles; captures admin + staff auth states |
| `frontend/e2e/mode-toggle.spec.ts` | Comment clarifying credentials used; auth state path unchanged |
| `reports/coverage-backend.html` | Regenerated HTML coverage report |

---

## 6. Zero Remaining Hardcoded Old Credentials

```
Searched: backend/tests/**  frontend/e2e/**  frontend/__tests__/**
Patterns: "admin" (as credential) | "Admin@123" | "staff01"

Result: 0 matches
```

All credential values are now injected via environment variables:

```
TEST_ADMIN_USERNAME  (default: budi_santoso)   role: admin → /dashboard
TEST_ADMIN_PASSWORD  (default: Admin@Dana2025!)
TEST_STAFF_USERNAME  (default: dewi_rahayu)    role: staff → /dashboard
TEST_STAFF_PASSWORD  (default: Staff@Dana2025!)
```
