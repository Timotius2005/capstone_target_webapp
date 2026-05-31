package app

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"pt-dana-sejahtera/internal/config"
	"pt-dana-sejahtera/internal/database"
	"pt-dana-sejahtera/internal/handlers"
	"pt-dana-sejahtera/internal/middleware"
	"pt-dana-sejahtera/internal/repository"
	"pt-dana-sejahtera/internal/security"
	"pt-dana-sejahtera/internal/services"
	"pt-dana-sejahtera/pkg/logger"
)

func Run() error {
	// Load .env (ignore error — env vars may be set externally)
	_ = godotenv.Load()

	// ── Logger ────────────────────────────────────────────────────────────────
	log := logger.New()
	defer func() { _ = log.Sync() }()

	// ── Config ────────────────────────────────────────────────────────────────
	cfg := config.New()

	// ── Runtime security mode — initialised from .env, changeable at runtime ──
	security.Init(cfg.SecurityMode)
	security.LogMode(log)

	// ── Startup info (no password logged) ─────────────────────────────────────
	log.Info("Starting PT. Dana Sejahtera backend",
		zap.String("environment", cfg.Environment),
		zap.String("security_mode", cfg.SecurityMode),
		zap.String("db_host", cfg.DBHost),
		zap.String("db_port", cfg.DBPort),
		zap.String("db_name", cfg.DBName),
		zap.String("db_user", cfg.DBUser),
		// NOTE: db_password is intentionally NOT logged
	)

	// ── Database — single initialization point ────────────────────────────────
	db, err := database.New(cfg.DatabaseURL, log)
	if err != nil {
		return fmt.Errorf("database connection: %w", err)
	}

	if err := database.AutoMigrate(db, log); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	// Load persisted mode from DB — overrides APP_SECURITY_MODE env var so that
	// mode changes made at runtime survive a container restart.
	database.LoadOrInitModeFromDB(db, log, cfg.SecurityMode)
	security.LogMode(log)

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	nasabahRepo := repository.NewNasabahRepository(db)
	loanRepo := repository.NewLoanRepository(db)
	txRepo := repository.NewTransactionRepository(db)

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := services.NewAuthService(userRepo, cfg.JWTSecret, log)
	nasabahSvc := services.NewNasabahService(nasabahRepo, userRepo, log)
	loanSvc := services.NewLoanService(loanRepo, nasabahRepo, log)
	txSvc := services.NewTransactionService(txRepo, loanRepo, nasabahRepo, log)
	extSvc := services.NewExternalService(log)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authH := handlers.NewAuthHandler(authSvc, log)
	nasabahH := handlers.NewNasabahHandler(nasabahSvc, log)
	loanH := handlers.NewLoanHandler(loanSvc, log)
	txH := handlers.NewTransactionHandler(txSvc, log)
	adminH := handlers.NewAdminHandler(userRepo, log)
	ssrfH := handlers.NewSSRFHandler(extSvc, log)
	configH := handlers.NewConfigHandler(log)
	systemH := handlers.NewSystemHandler(db, log)

	// ── Gin engine ────────────────────────────────────────────────────────────
	if security.IsVulnerable() {
		// TODO: Vulnerability Injection Point — OWASP API8 (Security Misconfiguration)
		// Debug mode leaks routing info and stack traces
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.HandleMethodNotAllowed = true // return 405 instead of 404 for wrong-method requests

	// ── Global middleware ─────────────────────────────────────────────────────
	r.Use(middleware.RequestLogger(log))
	r.Use(middleware.ErrorHandler(log))
	r.Use(middleware.CORS())
	r.Use(middleware.SecureHeaders())
	r.Use(middleware.RequestSizeLimit())

	// ── Runtime config endpoints (legacy — admin auth required) ──────────────
	// GET is public — frontend reads on every page load without auth.
	r.GET("/config/mode", configH.GetMode)

	// PUT requires a valid JWT and admin role. RoleCheck is used here instead of
	// AdminOnly because AdminOnly intentionally bypasses role enforcement in
	// sandbox mode (OWASP API5 demo). Config/mode must ALWAYS require admin.
	cfgRoutes := r.Group("/config")
	cfgRoutes.Use(middleware.AuthRequired(authSvc))
	cfgRoutes.Use(middleware.RoleCheck("admin"))
	{
		cfgRoutes.PUT("/mode", configH.SetMode)
	}

	// ── Public system mode API (no auth — optional LAB_KEY protection) ────────
	// These endpoints are intentionally unauthenticated for pentest lab use.
	// Enable LAB_KEY env var to restrict access to holders of the lab token.
	system := r.Group("/api/system")
	system.Use(middleware.LabKeyRequired())
	{
		system.GET("/mode", systemH.GetMode)
		system.PUT("/mode", systemH.SetMode)
		// Per-category OWASP vulnerability config (only active in vulnerable mode).
		system.GET("/vuln-config", systemH.GetVulnConfig)
		system.PUT("/vuln-config", systemH.SetVulnConfig)
	}

	// ── Health endpoints ──────────────────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":        "ok",
			"security_mode": security.GetMode(),
			"version":       "v1",
		})
	})

	r.GET("/health/db", func(c *gin.Context) {
		if err := database.HealthCheck(db); err != nil {
			log.Warn("DB health check failed", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"database": "disconnected",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	})

	// ── API v1 — current versioned API ────────────────────────────────────────
	v1 := r.Group("/api/v1")
	{
		// Public — login/register with login rate limit
		auth := v1.Group("/auth")
		auth.Use(middleware.LoginRateLimit(log))
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
			auth.POST("/refresh", authH.RefreshToken)
		}

		// Protected — require valid JWT
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired(authSvc))
		{
			protected.GET("/auth/me", authH.Me)

			// ── Nasabah ──────────────────────────────────────────────────────
			nas := protected.Group("/nasabah")
			nas.POST("", nasabahH.Register)       // nasabah registers own profile
			nas.GET("/me", nasabahH.GetMyProfile) // own profile shortcut
			nas.GET("", nasabahH.List)            // admin/staff: all; OWASP API4
			nas.GET("/:id", nasabahH.GetByID)     // OWASP API1 (BOLA)
			nas.PUT("/:id", nasabahH.Update)      // OWASP API1 + API3
			nas.DELETE("/:id", nasabahH.Delete)   // admin only in secure

			// ── Loans ────────────────────────────────────────────────────────
			loans := protected.Group("/loans")
			loans.POST("", loanH.Apply)                    // OWASP API4 + API6
			loans.GET("", loanH.List)                      // OWASP API1 + API4
			loans.GET("/:id", loanH.GetByID)               // OWASP API1 (BOLA)
			loans.PATCH("/:id/status", loanH.UpdateStatus) // OWASP API3 (BOPLA)
			loans.POST("/:id/approve", loanH.Approve)      // OWASP API5 (BFLA)
			loans.POST("/:id/reject", loanH.Reject)        // OWASP API5
			loans.GET("/:id/transactions", txH.ListByLoan) // OWASP API1

			// ── Transactions ─────────────────────────────────────────────────
			txs := protected.Group("/transactions")
			txs.GET("", txH.List)    // admin/staff only
			txs.POST("", txH.Create) // OWASP API5 + API6

			// ── Admin ─────────────────────────────────────────────────────────
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminOnly()) // OWASP API5: role guard in secure mode
			{
				admin.GET("/users", adminH.ListUsers)
				admin.PUT("/users/:id/role", adminH.UpdateRole)
				admin.GET("/stats", adminH.Stats)
			}

			// ── SSRF demo — OWASP API7 + API10 ───────────────────────────────
			protected.POST("/internal/fetch", ssrfH.Fetch)
		}
	}

	// ── API v0 — DEPRECATED: exposed only in vulnerable mode ──────────────────
	// TODO: Vulnerability Injection Point — OWASP API9 (Improper Inventory Management)
	if security.IsVulnerable() {
		v0 := r.Group("/api/v0")
		log.Warn("[VULNERABLE] Deprecated v0 routes registered — OWASP API9")
		{
			v0.GET("/loans", loanH.ListPublic)        // no auth, all loans
			v0.POST("/loans", loanH.ApplyPublic)      // no auth, no cap — OWASP API6
			v0.GET("/users", adminH.ListUsersPublic) // no auth, all users + hashes
			v0.GET("/debug", adminH.Debug)           // stack trace + internals
		}
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Info("Server ready",
		zap.String("addr", addr),
		zap.String("mode", security.GetMode()),
		zap.String("health", "GET /health"),
		zap.String("health_db", "GET /health/db"),
	)
	return r.Run(addr)
}
