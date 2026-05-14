# ──────────────────────────────────────────────────────────────────────────────
# PT. Dana Sejahtera — Test Orchestration Makefile
#
# Quick start:
#   make test              — run all tests + generate report
#   make test-backend      — Go unit + OWASP tests only
#   make test-frontend     — Jest unit tests only
#   make test-e2e          — Playwright e2e (requires running stack)
#   make test-security     — Python attack simulator (requires running backend)
#   make report            — regenerate the Markdown report from last run
#
# Prerequisites:
#   Backend: Go 1.23+, running PostgreSQL (or via docker-compose)
#   Frontend: Node.js 18+
#   Security: Python 3.11+
#   E2E: Playwright installed (npx playwright install chromium)
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: all test test-backend test-frontend test-e2e test-security report \
        deps-backend deps-frontend deps-security clean help

BACKEND_DIR   := backend
FRONTEND_DIR  := frontend
SECURITY_DIR  := tests/security
REPORTS_DIR   := reports

BACKEND_URL   ?= http://localhost:8080
FRONTEND_URL  ?= http://localhost:3000

# ── Colour helpers ─────────────────────────────────────────────────────────────
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
RESET  := \033[0m

# ── Default target ─────────────────────────────────────────────────────────────

all: test

# ──────────────────────────────────────────────────────────────────────────────
# MAIN: run all tests and generate final report
# Exit 0 only if secure-mode protections all pass.
# ──────────────────────────────────────────────────────────────────────────────

test: deps-backend test-backend test-frontend test-security report
	@echo ""
	@echo "$(GREEN)══════════════════════════════════════════════$(RESET)"
	@echo "$(GREEN)  Test suite complete — see reports/           $(RESET)"
	@echo "$(GREEN)══════════════════════════════════════════════$(RESET)"

# ──────────────────────────────────────────────────────────────────────────────
# BACKEND TESTS (Go)
# ──────────────────────────────────────────────────────────────────────────────

deps-backend:
	@echo "$(YELLOW)→ Syncing Go dependencies...$(RESET)"
	cd $(BACKEND_DIR) && go mod tidy

test-backend: deps-backend
	@echo "$(YELLOW)→ Running backend unit tests...$(RESET)"
	@mkdir -p $(REPORTS_DIR)
	cd $(BACKEND_DIR) && \
	  go test ./tests/unit/... -v -count=1 \
	    -coverprofile=../$(REPORTS_DIR)/backend-unit.coverage 2>&1 | tee ../$(REPORTS_DIR)/backend-unit.log
	@echo "$(YELLOW)→ Running OWASP backend tests...$(RESET)"
	cd $(BACKEND_DIR) && \
	  go test ./tests/owasp/... -v -count=1 \
	    -coverprofile=../$(REPORTS_DIR)/backend-owasp.coverage 2>&1 | tee ../$(REPORTS_DIR)/backend-owasp.log
	@echo "$(GREEN)✓ Backend tests complete$(RESET)"

test-backend-verbose:
	cd $(BACKEND_DIR) && go test ./tests/... -v -count=1

test-backend-race:
	@echo "$(YELLOW)→ Running backend tests with race detector...$(RESET)"
	cd $(BACKEND_DIR) && go test ./tests/... -race -count=1

# ──────────────────────────────────────────────────────────────────────────────
# FRONTEND TESTS (Jest + Playwright)
# ──────────────────────────────────────────────────────────────────────────────

deps-frontend:
	@echo "$(YELLOW)→ Installing frontend dependencies...$(RESET)"
	cd $(FRONTEND_DIR) && npm install --silent

test-frontend: deps-frontend
	@echo "$(YELLOW)→ Running frontend Jest unit tests...$(RESET)"
	@mkdir -p $(REPORTS_DIR)
	cd $(FRONTEND_DIR) && \
	  npx jest --coverage --passWithNoTests \
	    --coverageDirectory=../$(REPORTS_DIR)/frontend-coverage 2>&1 | tee ../$(REPORTS_DIR)/frontend-unit.log
	@echo "$(GREEN)✓ Frontend unit tests complete$(RESET)"

test-e2e: deps-frontend
	@echo "$(YELLOW)→ Running Playwright e2e tests...$(RESET)"
	@echo "$(YELLOW)   Requires: frontend at $(FRONTEND_URL), backend at $(BACKEND_URL)$(RESET)"
	cd $(FRONTEND_DIR) && \
	  PLAYWRIGHT_BASE_URL=$(FRONTEND_URL) \
	  PLAYWRIGHT_API_URL=$(BACKEND_URL) \
	  npx playwright test 2>&1 | tee ../$(REPORTS_DIR)/e2e.log
	@echo "$(GREEN)✓ E2E tests complete$(RESET)"

install-playwright:
	cd $(FRONTEND_DIR) && npx playwright install chromium

# ──────────────────────────────────────────────────────────────────────────────
# SECURITY TESTS (Python attack simulator)
# ──────────────────────────────────────────────────────────────────────────────

deps-security:
	@echo "$(YELLOW)→ Installing Python dependencies...$(RESET)"
	pip install -r $(SECURITY_DIR)/requirements.txt -q

test-security: deps-security
	@echo "$(YELLOW)→ Running OWASP attack simulator...$(RESET)"
	@echo "$(YELLOW)   Target: $(BACKEND_URL)$(RESET)"
	@mkdir -p $(REPORTS_DIR)
	python $(SECURITY_DIR)/attack_simulator.py \
	  --base-url $(BACKEND_URL) \
	  --report $(REPORTS_DIR)/security-validation.json 2>&1 | tee $(REPORTS_DIR)/security.log; \
	EXIT=$$?; \
	if [ $$EXIT -eq 0 ]; then \
	  echo "$(GREEN)✓ Security tests: all secure-mode protections PASSED$(RESET)"; \
	elif [ $$EXIT -eq 2 ]; then \
	  echo "$(RED)✗ Security tests: backend unreachable — skipping$(RESET)"; \
	else \
	  echo "$(RED)✗ Security tests: secure-mode protections FAILED$(RESET)"; \
	  exit $$EXIT; \
	fi

# ──────────────────────────────────────────────────────────────────────────────
# REPORT GENERATION
# ──────────────────────────────────────────────────────────────────────────────

report:
	@echo "$(YELLOW)→ Generating test report...$(RESET)"
	@mkdir -p $(REPORTS_DIR)
	@bash scripts/generate_report.sh
	@echo "$(GREEN)✓ Report saved to $(REPORTS_DIR)/test-report.md$(RESET)"

# ──────────────────────────────────────────────────────────────────────────────
# UTILITIES
# ──────────────────────────────────────────────────────────────────────────────

clean:
	@echo "$(YELLOW)→ Cleaning report artefacts...$(RESET)"
	rm -f $(REPORTS_DIR)/*.log
	rm -f $(REPORTS_DIR)/*.coverage
	rm -f $(REPORTS_DIR)/*.json
	rm -f $(REPORTS_DIR)/test-report.md
	rm -rf $(REPORTS_DIR)/frontend-coverage
	rm -rf $(REPORTS_DIR)/playwright-report
	@echo "$(GREEN)✓ Clean$(RESET)"

help:
	@echo ""
	@echo "PT. Dana Sejahtera — Test Targets"
	@echo "──────────────────────────────────"
	@echo "  make test              Full suite (backend + frontend + security + report)"
	@echo "  make test-backend      Go unit & OWASP tests"
	@echo "  make test-frontend     Jest unit tests"
	@echo "  make test-e2e          Playwright e2e (stack must be running)"
	@echo "  make test-security     Python OWASP attack simulator"
	@echo "  make report            Regenerate Markdown report"
	@echo "  make clean             Remove generated artefacts"
	@echo ""
	@echo "  BACKEND_URL=$(BACKEND_URL)"
	@echo "  FRONTEND_URL=$(FRONTEND_URL)"
	@echo ""
