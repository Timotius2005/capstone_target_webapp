package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"pt-dana-sejahtera/internal/config"
	"pt-dana-sejahtera/internal/database"
	"pt-dana-sejahtera/internal/handler"
	"pt-dana-sejahtera/internal/middleware"
	"pt-dana-sejahtera/internal/repository"
	"pt-dana-sejahtera/internal/usecase"
)

func Run() error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize configuration
	cfg := config.New()

	// Initialize database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	nasabahRepo := repository.NewNasabahRepository(db)
	loanRepo := repository.NewLoanRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg.JWTSecret)
	nasabahUsecase := usecase.NewNasabahUsecase(nasabahRepo)
	loanUsecase := usecase.NewLoanUsecase(loanRepo, nasabahRepo)
	transactionUsecase := usecase.NewTransactionUsecase(transactionRepo, loanRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUsecase)
	nasabahHandler := handler.NewNasabahHandler(nasabahUsecase)
	loanHandler := handler.NewLoanHandler(loanUsecase)
	transactionHandler := handler.NewTransactionHandler(transactionUsecase)

	// Initialize Gin router
	router := gin.Default()

	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			// User routes
			users := protected.Group("/users")
			users.GET("/profile", authHandler.GetProfile)

			// Nasabah routes
			nasabah := protected.Group("/nasabah")
			nasabah.GET("", nasabahHandler.List)
			nasabah.POST("", nasabahHandler.Create)
			nasabah.GET("/:id", nasabahHandler.GetByID)
			nasabah.PUT("/:id", nasabahHandler.Update)
			nasabah.DELETE("/:id", nasabahHandler.Delete)

			// Loan routes
			loans := protected.Group("/loans")
			loans.GET("", loanHandler.List)
			loans.POST("", loanHandler.Create)
			loans.GET("/:id", loanHandler.GetByID)
			loans.PUT("/:id", loanHandler.Update)
			loans.DELETE("/:id", loanHandler.Delete)

			// Transaction routes
			transactions := protected.Group("/transactions")
			transactions.GET("", transactionHandler.List)
			transactions.POST("", transactionHandler.Create)
			transactions.GET("/:id", transactionHandler.GetByID)
		}
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)
	return router.Run(addr)
}