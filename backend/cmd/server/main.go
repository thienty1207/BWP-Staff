package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/thienty1207/BWP-Staff/backend/app"
	"github.com/thienty1207/BWP-Staff/backend/config"
	"github.com/thienty1207/BWP-Staff/backend/shared"
)

func main() {
	_ = godotenv.Load()

	settings, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	pool, err := shared.Connect(context.Background(), settings)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := shared.RunMigrations(context.Background(), pool, "migrations", time.Duration(settings.DatabaseAcquireTimeoutSeconds)*time.Second); err != nil {
		log.Fatal(err)
	}

	server := app.New()
	if err := server.Listen("127.0.0.1:3000"); err != nil {
		log.Fatal(err)
	}
}
