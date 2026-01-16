package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/gin-contrib/cors"
	"github.com/sisghe/inventory-management-system/backend/db"
	"github.com/sisghe/inventory-management-system/backend/routes"
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

	// Verifica reale connessione DB con Ping (pgxpool)
	{
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := db.Pool.Ping(ctx); err != nil {
			log.Fatal("❌ DB Ping failed:", err)
		}
	}

	log.Println(" Connected to database successfully (pgxpool ping OK)!")

	// Server Gin
	r := gin.Default()

	r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
	}))
	// Registra tutte le rotte (pubbliche + protette)
	routes.Register(r)

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
