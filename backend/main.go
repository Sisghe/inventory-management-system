package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/sisghe/inventory-management-system/backend/db"
)

func main() {
	// Carica variabili d'ambiente
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found (continuo comunque usando env di sistema)")
	}

	// Connessione al database (pgxpool)
	if err := db.Connect(); err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer db.Pool.Close()

	// ✅ Verifica reale connessione DB con Ping (pgxpool)
	{
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := db.Pool.Ping(ctx); err != nil {
			log.Fatal("❌ DB Ping failed:", err)
		}
	}

	log.Println("✅ Connected to database successfully (pgxpool ping OK)!")

	// Server Gin
	r := gin.Default()

	// Healthcheck base
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status":  "server up",
		})
	})

	// Healthcheck DB (fa query vera)
	r.GET("/db-ping", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		var one int
		err := db.Pool.QueryRow(ctx, "SELECT 1").Scan(&one)
		if err != nil {
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

	// Porta server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
