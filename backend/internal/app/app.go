package app

import (
	"fmt"

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

	// ── Security mode banner ──────────────────────────────────────────────────
	security.LogMode(log)

	// ── Config ────────────────────────────────────────────────────────────────
	cfg := config.New()

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database connection: %w", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	log.Info("Database connected and migrated")

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

	// ── Gin engine ────────────────────────────────────────────────────────────
	if security.IsVulnerable() {
		// TODO: Vulnerability Injection Point — OWASP API8 (Security Misconfiguration)
		// Debug mode leaks routing info and stack traces
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// ── Global middleware ─────────────────────────────────────────────────────
	r.Use(middleware.RequestLogger(log))
	r.Use(middleware.ErrorHandler(log))
	r.Use(middleware.CORS())
	r.Use(middleware.SecureHeaders())
	r.Use(middleware.RequestSizeLimit())

	// ── Health ────────────────────────────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":        "ok",
			"security_mode": security.GetMode(),
			"version":       "v1",
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
			loans.POST("", loanH.Apply)                      // OWASP API4 + API6
			loans.GET("", loanH.List)                        // OWASP API1 + API4
			loans.GET("/:id", loanH.GetByID)                 // OWASP API1 (BOLA)
			loans.PATCH("/:id/status", loanH.UpdateStatus)   // OWASP API3 (BOPLA)
			loans.POST("/:id/approve", loanH.Approve)        // OWASP API5 (BFLA)
			loans.POST("/:id/reject", loanH.Reject)          // OWASP API5
			loans.GET("/:id/transactions", txH.ListByLoan)   // OWASP API1

			// ── Transactions ─────────────────────────────────────────────────
			txs := protected.Group("/transactions")
			txs.GET("", txH.List)       // admin/staff only
			txs.POST("", txH.Create)    // OWASP API5 + API6

			// ── Admin ─────────────────────────────────────────────────────────
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminOnly()) // OWASP API5: role guard in secure mode
			{
				admin.GET("/users", adminH.ListUsers)
				admin.PUT("/users/:id/role", adminH.UpdateRole)
				admin.GET("/stats", adminH.Stats)
			}

			// ── SSRF demo ─────────────────────────────────────────────────────
			// OWASP API7 + API10
			protected.POST("/internal/fetch", ssrfH.Fetch)
		}
	}

	// ── API v0 — DEPRECATED: exposed only in vulnerable mode ──────────────────
	// TODO: Vulnerability Injection Point — OWASP API9 (Improper Inventory Management)
	// These endpoints have no auth and expose everything
	if security.IsVulnerable() {
		v0 := r.Group("/api/v0")
		log.Warn("[VULNERABLE] Deprecated v0 routes registered — OWASP API9")
		{
			v0.GET("/loans", loanH.ListPublic)           // no auth, all loans
			v0.GET("/users", adminH.ListUsersPublic)     // no auth, all users + hashes
			v0.GET("/debug", adminH.Debug)               // stack trace + internals
		}
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Info("Server starting",
		zap.String("addr", addr),
		zap.String("mode", security.GetMode()),
	)
	return r.Run(addr)
}
