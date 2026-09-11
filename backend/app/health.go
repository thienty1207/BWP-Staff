package app

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func healthHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func readinessHandler(pool *pgxpool.Pool, timeoutSeconds int) fiber.Handler {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 1
	}
	readinessTimeout := time.Duration(timeoutSeconds) * time.Second

	return func(c fiber.Ctx) error {
		if pool == nil {
			err := errors.New("database pool is not configured")
			logReadinessFailure(c, err)
			return &AppError{
				Code:       "not_ready",
				Message:    "Service not ready",
				HTTPStatus: fiber.StatusServiceUnavailable,
			}
		}

		ctx, cancel := context.WithTimeout(c.Context(), readinessTimeout)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			logReadinessFailure(c, err)
			return &AppError{
				Code:       "not_ready",
				Message:    "Service not ready",
				HTTPStatus: fiber.StatusServiceUnavailable,
			}
		}

		return c.JSON(fiber.Map{"status": "ready"})
	}
}

func logReadinessFailure(c fiber.Ctx, err error) {
	log.Printf("request_id=%s readiness_error=%v", requestid.FromContext(c), err)
}
