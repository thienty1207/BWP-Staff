package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/thienty1207/BWP-Staff/backend/admin"
	"github.com/thienty1207/BWP-Staff/backend/config"
	"github.com/thienty1207/BWP-Staff/backend/shared"
)

func main() {
	_ = godotenv.Load()

	settings, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if !settings.SeedDevelopmentData || settings.AppEnv != "development" {
		log.Fatal("development seed requires APP_ENV=development and SEED_DEVELOPMENT_DATA=true")
	}

	pool, err := shared.Connect(context.Background(), settings)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := shared.RunMigrations(context.Background(), pool, "migrations", time.Duration(settings.DatabaseAcquireTimeoutSeconds)*time.Second); err != nil {
		log.Fatal(err)
	}
	if err := admin.SeedDevelopmentFixtures(context.Background(), pool); err != nil {
		log.Fatal(err)
	}
	outcome, err := admin.SeedDevelopmentAdmin(context.Background(), pool, settings.SeedAdmin)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("development admin seed: %s\n", outcome)
}
