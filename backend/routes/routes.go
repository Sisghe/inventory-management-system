package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sisghe/inventory-management-system/backend/db"
	"github.com/sisghe/inventory-management-system/backend/handlers"
	"github.com/sisghe/inventory-management-system/backend/middleware"
	"github.com/sisghe/inventory-management-system/backend/repositories"
	"github.com/sisghe/inventory-management-system/backend/services"
)

// Register registra tutte le rotte dell'app su Gin.
func Register(r *gin.Engine) {
	// ====== Dependency wiring (DI manuale) ======

	// repositories
	userRepo := repositories.NewUserRepository()
	productRepo := repositories.NewProductRepository()
	productTypeRepo := repositories.NewProductTypeRepository()

	// services
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	productService := services.NewProductService(productRepo, productTypeRepo)

	// handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
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
		// login reale
		auth.POST("/login", authHandler.Login)

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
		// endpoint test: verifica token + middleware
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

		// ====================================================
		// ====== 🔽 NUOVE ROTTE AGGIUNTE (PRODOTTI) 🔽 ======
		// ====================================================

		// Tipi prodotto (Buste, Carta, Toner)
		// Usata dal frontend per dropdown / select
		api.GET("/product-types", productTypeHandler.List)

		// CRUD prodotti
		api.GET("/products", productHandler.List)
		api.POST("/products", productHandler.Create)
		api.PUT("/products/:id", productHandler.Update)
		api.DELETE("/products/:id", productHandler.Delete)
	}
}
