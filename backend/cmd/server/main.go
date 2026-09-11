package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/thienty1207/BWP-Staff/backend/app"
	"github.com/thienty1207/BWP-Staff/backend/config"
	"github.com/thienty1207/BWP-Staff/backend/shared"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(parent context.Context) error {
	if err := loadDotEnv(); err != nil {
		return err
	}

	settings, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	serverContext, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := shared.Connect(serverContext, settings)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	if err := shared.RunMigrations(
		serverContext,
		pool,
		"migrations",
		time.Duration(settings.DatabaseAcquireTimeoutSeconds)*time.Second,
	); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	server := app.New(app.AppState{DB: pool}, settings)
	return serve(serverContext, server, settings)
}

func loadDotEnv() error {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load .env: %w", err)
	}
	return nil
}

func serve(ctx context.Context, server *fiber.App, settings config.Config) error {
	listenerErrors := make(chan error, 1)
	log.Printf("server starting bind_address=%s", settings.BackendBindAddress)
	go func() {
		listenerErrors <- server.Listen(settings.BackendBindAddress)
	}()

	select {
	case err := <-listenerErrors:
		if ctx.Err() != nil {
			return nil
		}
		if err == nil {
			return errors.New("HTTP listener stopped unexpectedly")
		}
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
		log.Printf("shutdown initiated")
		shutdownContext, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(settings.BackendShutdownTimeoutSeconds)*time.Second,
		)
		defer cancel()

		if err := server.ShutdownWithContext(shutdownContext); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		log.Printf("shutdown completed")
		return nil
	}
}
