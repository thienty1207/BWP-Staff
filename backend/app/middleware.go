package app

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func serverRequestID() fiber.Handler {
	requestID := requestid.New()
	return func(c fiber.Ctx) error {
		c.Request().Header.Del(fiber.HeaderXRequestID)
		return requestID(c)
	}
}

func requestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		startedAt := time.Now()
		err := c.Next()
		status := c.Response().StatusCode()
		if err != nil {
			status = statusForError(err)
		}

		log.Printf(
			"request_id=%s method=%s path=%s status=%d latency=%s",
			requestid.FromContext(c),
			c.Method(),
			c.Path(),
			status,
			time.Since(startedAt),
		)
		return err
	}
}
