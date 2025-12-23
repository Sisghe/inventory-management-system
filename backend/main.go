package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Carica .env dalla root
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found")
	}

	// Stringa di connessione da .env
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	// Connessione al DB
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer conn.Close(context.Background())

	fmt.Println("Connected to database inventory_db successfully!")

	// Server Gin
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "pong",
			"db_status": "connected",
		})
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ":8080"
	}
	fmt.Printf("Server running on %s\n", port)
	r.Run(port)
}
