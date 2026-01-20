package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/sisghe/inventory-management-system/backend/db"
	"github.com/sisghe/inventory-management-system/backend/routes"
)

func main() {
	// Carica variabili d'ambiente (non bloccante).
	// Best practice: prova prima .env nella working dir, poi fallback a ../.env.
	if err := godotenv.Load(); err != nil {
		if err2 := godotenv.Load("../.env"); err2 != nil {
			log.Println("No .env file found (continuo comunque usando env di sistema)")
		}
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
			log.Fatal("DB Ping failed:", err)
		}
	}

	log.Println("Connected to database successfully (pgxpool ping OK)!")

	// Server Gin
	// Nota: gin.Default() include già Logger + Recovery
	r := gin.Default()

	// Best practice: evita warning "You trusted all proxies".
	// In locale va benissimo nil. In prod potrai impostare IP proxy reali.
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Println("SetTrustedProxies warning:", err)
	}

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Accept",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Registra tutte le rotte (pubbliche + protette)
	routes.Register(r)

	// Best practice: supporta anche PORT (utile su molte piattaforme), fallback a SERVER_PORT.
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("SERVER_PORT")
	}
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
