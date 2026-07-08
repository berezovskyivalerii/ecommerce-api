package main

import (
	"log"
	"os"

	"github.com/berezovskyivalerii/ecommerce-api/internal/app"
	"github.com/berezovskyivalerii/ecommerce-api/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on environment variables")
	}
	dbURL := os.Getenv("DB_URL")

	cfg := config.New(dbURL)

	application := app.New(cfg)
	if err := application.Run(); err != nil {
		application.Close()
		log.Fatalf("server stopped: %v", err)
	}
}
