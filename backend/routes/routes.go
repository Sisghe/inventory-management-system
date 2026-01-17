package routes

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sisghe/inventory-management-system/backend/db"
	"github.com/sisghe/inventory-management-system/backend/handlers"
	"github.com/sisghe/inventory-management-system/backend/middleware"
	"github.com/sisghe/inventory-management-system/backend/repositories"
	"github.com/sisghe/inventory-management-system/backend/services"
	"github.com/sisghe/inventory-management-system/backend/utils"
)

// Register registra tutte le rotte dell'app su Gin.
func Register(r *gin.Engine) {
	// ====== Dependency wiring (DI manuale) ======

	// repositories
	userRepo := repositories.NewUserRepository()
	productRepo := repositories.NewProductRepository()
	productTypeRepo := repositories.NewProductTypeRepository()
	emailVerRepo := repositories.NewEmailVerificationRepository()

	// mailer (SMTP / Mailtrap)
	mailer, err := utils.NewSMTPMailerFromEnv()
	if err != nil {
		// in dev puoi anche decidere di fatal, ma così almeno il backend parte
		// e ti segnala che la verifica email non potrà inviare mail finché non setti env.
		log.Println("SMTP mailer not configured:", err)
		mailer = nil
	}

	// services
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	productService := services.NewProductService(productRepo, productTypeRepo)

	var emailVerifyService *services.EmailVerificationService
	if mailer != nil {
		emailVerifyService = services.NewEmailVerificationService(emailVerRepo, userRepo, mailer)
	} else {
		emailVerifyService = nil
	}

	// handlers
	authHandler := handlers.NewAuthHandler(authService, emailVerifyService)
	userHandler := handlers.NewUserHandler(userService, emailVerifyService)
	productHandler := handlers.NewProductHandler(productService)
	productTypeHandler := handlers.NewProductTypeHandler(productTypeRepo)

	// ====== Rotte pubbliche (accessibili senza token) ======
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Healthcheck DB (pubblico)
	r.GET("/db-ping", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		var one int
		if err := db.Pool.QueryRow(ctx, "SELECT 1").Scan(&one); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"db":    "down",
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"db":     "up",
			"select": one,
		})
	})

	// ====== Gruppo /auth (pubblico) ======
	auth := r.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authHandler.Logout)

		// ✅ verifica email
		auth.POST("/verify-email", authHandler.VerifyEmail)

		// Placeholder: recupero password
		auth.POST("/forgot-password", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
		})

		// Placeholder: reset password
		auth.POST("/reset-password", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
		})
	}

	// ====== Rotte protette (richiedono token) ======
	api := r.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		api.GET("/me", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"user_id":  c.GetInt("user_id"),
				"username": c.GetString("username"),
			})
		})

		// ====== CRUD UTENTI ======
		api.GET("/users", userHandler.List)
		api.POST("/users", userHandler.Create)
		api.PUT("/users/:id", userHandler.Update)
		api.DELETE("/users/:id", userHandler.Delete)

		// Tipi prodotto
		api.GET("/product-types", productTypeHandler.List)

		// CRUD prodotti
		api.GET("/products", productHandler.List)
		api.POST("/products", productHandler.Create)
		api.PUT("/products/:id", productHandler.Update)
		api.DELETE("/products/:id", productHandler.Delete)
	}
}
