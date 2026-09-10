package app

import "github.com/gofiber/fiber/v3"

// New builds the HTTP application without opening a listener.
func New() *fiber.App {
	app := fiber.New()
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	return app
}
