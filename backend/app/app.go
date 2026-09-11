package app

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/thienty1207/BWP-Staff/backend/client/auth"
	"github.com/thienty1207/BWP-Staff/backend/config"
)

// New builds the HTTP application without opening a listener or a database
// connection.
func New(state AppState, settings config.Config) *fiber.App {
	server := fiber.New(fiber.Config{
		ErrorHandler: handleError,
	})

	server.Use(serverRequestID())
	server.Use(requestLogger())
	server.Use(recover.New())
	server.Use(corsMiddleware(settings.FrontendOrigin))

	server.Get("/health", healthHandler)
	server.Get("/ready", readinessHandler(state.DB, settings.DatabaseAcquireTimeoutSeconds))
	api := server.Group("/api/v1")
	auth.RegisterRoutes(api, state.DB, settings)

	return server
}

func corsMiddleware(frontendOrigin string) fiber.Handler {
	settings := cors.Config{
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodHead,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		},
		AllowHeaders: []string{
			fiber.HeaderContentType,
			fiber.HeaderAccept,
		},
		ExposeHeaders:    []string{fiber.HeaderXRequestID},
		AllowCredentials: true,
	}

	if frontendOrigin == "" {
		settings.AllowOriginsFunc = func(string) bool { return false }
	} else {
		settings.AllowOrigins = []string{frontendOrigin}
	}

	return cors.New(settings)
}
