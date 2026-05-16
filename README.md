# PT. Dana Sejahtera — Fintech Security Lab

A dual-mode fintech loan-management system built for security education, CTF competitions, and API penetration-testing training.

The application runs in **two modes** that can be switched at runtime without a container restart:

- **Secure mode** — all OWASP API Security Top 10 protections are active
- **Vulnerable (sandbox) mode** — intentional weaknesses are exposed, individually configurable per OWASP category

---

## Tech Stack

| Layer       | Technology                                  |
|-------------|---------------------------------------------|
| Backend     | Go 1.21 · Gin · GORM · JWT · bcrypt · Zap  |
| Frontend    | Next.js 14 · TypeScript · Tailwind CSS      |
| Database    | PostgreSQL 15                               |
| Container   | Docker · Docker Compose                     |
| Testing     | Go unit/OWASP tests · Jest · Playwright E2E |

---

## Quick Start (Docker)

```bash
# 1. Clone
git clone <repo-url> && cd capstone_target_webapp

# 2. Start all services (postgres → backend → frontend)
docker compose up --build

# 3. Open the app
#    Frontend : http://localhost:3000
#    Backend  : http://localhost:8080
#    Health   : http://localhost:8080/health
```

> First run pulls images and builds containers — allow 3–5 minutes.
> Subsequent runs are instant unless source files changed.

---

## Default Credentials

| Username      | Password           | Role    |
|---------------|--------------------|---------|
| `admin`       | `Admin@123`        | admin   |
| `staff01`     | `Admin@123`        | staff   |
| `budi_santoso`| `Admin@Dana2025!`  | admin   |
| `dewi_rahayu` | `Staff@Dana2025!`  | staff   |

> **admin / budi_santoso** can approve loans and access `/admin/*`.
> **staff01 / dewi_rahayu** can process but not approve.

---

## Environment Variables

Copy the template and adjust if needed:

```bash
cp .env.example .env
```

| Variable           | Default                                    | Description                                  |
|--------------------|--------------------------------------------|----------------------------------------------|
| `DB_USER`          | `fintech`                                  | PostgreSQL username                          |
| `DB_PASSWORD`      | `securepass`                               | PostgreSQL password                          |
| `DB_NAME`          | `nasabahdb`                               | PostgreSQL database name                     |
| `JWT_SECRET`       | `change-me-in-production-min-32-chars!!`   | HS256 signing key — change before production |
| `APP_SECURITY_MODE`| `secure`                                   | Initial mode (`secure` or `vulnerable`)      |
| `LAB_KEY`          | *(empty)*                                  | Optional token for `/api/system/*` endpoints |

> The application reads the persisted mode from the database on startup — `APP_SECURITY_MODE` only sets the initial value on first boot.

---

## Secure Mode vs Vulnerable Mode

| Behaviour                    | Secure                          | Vulnerable (Sandbox)                    |
|------------------------------|---------------------------------|-----------------------------------------|
| JWT storage                  | `sessionStorage` (tab-scoped)   | `localStorage` (persistent, XSS risk)  |
| Token expiry                 | 15 minutes + refresh            | 100 years (no expiry)                   |
| Password comparison          | `bcrypt.CompareHashAndPassword` | Plain-text string compare               |
| Account lockout              | 5 failed attempts → locked      | No lockout                              |
| Rate limiting                | 5 req/min on `/auth/login`      | Disabled                                |
| CORS                         | Whitelist only                  | Wildcard `*`                            |
| Security headers             | Full OWASP headers set          | Server version + debug headers exposed  |
| Admin role check             | Enforced                        | Bypassed (any JWT reaches admin routes) |
| Error detail                 | Generic messages                | Full stack trace in response            |
| GORM query logging           | Silent                          | Debug (raw SQL with parameters)         |
| Deprecated v0 API routes     | Not registered                  | Registered (no auth, full data dump)    |
| SSRF allowlist               | Domain whitelist enforced       | Fetch any URL including metadata        |

### Switching modes

**Via the UI** — Click the fixed-position mode switcher (bottom-right corner, visible on every page).

**Via API** — No authentication required:

```bash
# Switch to vulnerable
curl -X PUT http://localhost:8080/api/system/mode \
     -H 'Content-Type: application/json' \
     -d '{"mode":"vulnerable"}'

# Switch back to secure
curl -X PUT http://localhost:8080/api/system/mode \
     -d '{"mode":"secure"}' -H 'Content-Type: application/json'
```

The mode is persisted in the database — a container restart does **not** reset it.

---

## Per-Category OWASP Toggle (Fine-Grained Control)

In vulnerable mode, each OWASP Top 10 category can be independently enabled or disabled without switching the entire mode:

```bash
# Get current config
curl http://localhost:8080/api/system/vuln-config

# Disable A03 Injection while keeping everything else vulnerable
curl -X PUT http://localhost:8080/api/system/vuln-config \
     -H 'Content-Type: application/json' \
     -d '{
       "A01_BrokenAccessControl": true,
       "A02_CryptographicFailures": true,
       "A03_Injection": false,
       "A04_InsecureDesign": true,
       "A05_SecurityMisconfiguration": true,
       "A06_VulnerableComponents": true,
       "A07_AuthenticationFailures": true,
       "A08_SoftwareDataIntegrityFailures": true,
       "A09_SecurityLoggingFailures": true,
       "A10_SSRF": true
     }'
```

The **VulnConfigPanel** component in the frontend admin panel provides a graphical toggle for each category.

---

## OWASP API Security Top 10 Coverage

| ID   | Category                           | Secure Behaviour                               | Vulnerable Behaviour                                    |
|------|------------------------------------|------------------------------------------------|---------------------------------------------------------|
| A01  | Broken Object Level Authorization  | Ownership checked on every resource            | Any user can read/modify any record                     |
| A02  | Cryptographic Failures             | Password hash hidden; NIK masked               | `password_hash`, raw NIK returned in responses          |
| A03  | Injection / BOPLA                  | Parameterised queries; NIK uniqueness enforced | NIK bypass; unchecked mass-assignment writes            |
| A04  | Insecure Design                    | Pagination limits; pending-loan cap enforced   | Full table dumps; unlimited loan creation               |
| A05  | Security Misconfiguration          | Silent GORM logging; no debug headers          | Raw SQL logged with params; server version exposed      |
| A06  | Vulnerable Components              | Loan amount & business rules validated         | Amount validation and business-rule enforcement removed |
| A07  | Authentication Failures            | bcrypt; 15 min JWT; lockout after 5 attempts   | Plain-text compare; 100-year JWT; no lockout            |
| A08  | Software & Data Integrity Failures | Staff cannot approve loans (admin-only)        | Staff can set loan status to `approved`                 |
| A09  | Security Logging & Monitoring      | Rate limit on login (brute-force detected)     | Rate limit disabled; brute-force undetected             |
| A10  | SSRF                               | Domain allowlist + private-IP block            | Fetch any URL (metadata endpoints reachable)            |

---

## API Reference

### Authentication
| Method | Endpoint               | Auth | Description                  |
|--------|------------------------|------|------------------------------|
| POST   | `/api/v1/auth/register`| —    | Register new user            |
| POST   | `/api/v1/auth/login`   | —    | Login → returns JWT          |
| POST   | `/api/v1/auth/refresh` | —    | Refresh JWT                  |
| GET    | `/api/v1/auth/me`      | JWT  | Current user info            |

### System / Mode
| Method | Endpoint                  | Auth | Description                        |
|--------|---------------------------|------|------------------------------------|
| GET    | `/api/system/mode`        | —    | Get current mode                   |
| PUT    | `/api/system/mode`        | —    | Switch mode (persist to DB)        |
| GET    | `/api/system/vuln-config` | —    | Get per-category OWASP config      |
| PUT    | `/api/system/vuln-config` | —    | Update per-category OWASP config   |

### Nasabah (Customer)
| Method | Endpoint            | Auth | OWASP |
|--------|---------------------|------|-------|
| POST   | `/api/v1/nasabah`   | JWT  | A03   |
| GET    | `/api/v1/nasabah/me`| JWT  | —     |
| GET    | `/api/v1/nasabah`   | JWT  | A04   |
| GET    | `/api/v1/nasabah/:id`| JWT | A01   |
| PUT    | `/api/v1/nasabah/:id`| JWT | A01 A03 |
| DELETE | `/api/v1/nasabah/:id`| JWT | A01   |

### Loans
| Method | Endpoint                    | Auth | OWASP  |
|--------|-----------------------------|------|--------|
| POST   | `/api/v1/loans`             | JWT  | A04 A06|
| GET    | `/api/v1/loans`             | JWT  | A01 A04|
| GET    | `/api/v1/loans/:id`         | JWT  | A01    |
| PATCH  | `/api/v1/loans/:id/status`  | JWT  | A03    |
| POST   | `/api/v1/loans/:id/approve` | JWT  | A05 A07|
| POST   | `/api/v1/loans/:id/reject`  | JWT  | A05    |

### Admin
| Method | Endpoint                  | Auth       | OWASP |
|--------|---------------------------|------------|-------|
| GET    | `/api/v1/admin/users`     | JWT + admin| A02 A05|
| PUT    | `/api/v1/admin/users/:id/role`| JWT + admin| A03 |
| GET    | `/api/v1/admin/stats`     | JWT + admin| —     |

### Deprecated v0 (vulnerable mode only — OWASP A09)
| Method | Endpoint        | Auth | Description                         |
|--------|-----------------|------|-------------------------------------|
| GET    | `/api/v0/loans` | —    | All loans, no auth                  |
| GET    | `/api/v0/users` | —    | All users + password hashes, no auth|
| GET    | `/api/v0/debug` | —    | Stack trace + internals             |

---

## Local Development (without Docker)

### Backend

```bash
cp backend/.env.example backend/.env
# Edit DB_HOST to point at your postgres instance
cd backend && go run ./cmd
```

### Frontend

```bash
cp frontend/.env.local.example frontend/.env.local
cd frontend && npm install && npm run dev
```

### Running tests

```bash
# Backend unit + OWASP tests
cd backend && go test ./...

# Frontend unit tests
cd frontend && npm test

# E2E (requires running stack)
cd frontend && npx playwright test
```

---

## Project Structure

```
capstone_target_webapp/
├── backend/
│   ├── cmd/main.go              # Entry point
│   ├── internal/
│   │   ├── app/app.go           # Router + middleware wiring
│   │   ├── config/              # Environment config
│   │   ├── database/            # GORM init + AutoMigrate + mode persistence
│   │   ├── handlers/            # HTTP handlers (one per domain)
│   │   ├── middleware/          # Auth, CORS, rate-limit, security headers
│   │   ├── models/              # GORM models + DTOs
│   │   ├── repository/          # Data access layer
│   │   ├── security/            # Runtime mode + per-category vuln config
│   │   └── services/            # Business logic
│   ├── migrations/              # Raw SQL migration files
│   ├── pkg/
│   │   ├── logger/              # Zap logger factory
│   │   └── seed/                # Test-user seeder
│   ├── tests/
│   │   ├── owasp/               # OWASP Top 10 integration tests
│   │   └── unit/                # Unit tests (auth, nasabah, loans)
│   ├── Dockerfile
│   ├── .env.example
│   └── go.mod
├── frontend/
│   ├── app/                     # Next.js 14 App Router pages
│   ├── components/              # React components (Navbar, ModeBadge, etc.)
│   ├── contexts/ModeContext.tsx # Global mode state + vuln config
│   ├── services/                # Axios API client + auth service
│   ├── utils/securityMode.ts    # Runtime mode module (non-React)
│   ├── middleware.ts             # Next.js edge middleware (auth guard)
│   ├── Dockerfile
│   └── next.config.js
├── database/
│   └── schema.sql               # Postgres init script (tables + seed data)
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## Security Disclaimer

This system contains **intentional vulnerabilities** for educational purposes.

- Deploy only on isolated lab networks or localhost
- Never expose the vulnerable mode to the public internet
- Default credentials are documented above — change them before any shared deployment
- The `LAB_KEY` env var can restrict mode-switch access to holders of that token

---

## License

For educational and research use only.
