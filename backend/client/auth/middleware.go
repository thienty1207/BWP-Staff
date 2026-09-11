package auth

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/thienty1207/BWP-Staff/backend/shared/httperror"
)

type principalContextKey struct{}

func (service *Service) RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		principal, err := service.Authenticate(c.Context(), c.Cookies(sessionCookieName))
		if errors.Is(err, ErrUnauthenticated) {
			return unauthenticatedError()
		}
		if err != nil {
			return err
		}
		c.Locals(principalContextKey{}, principal)
		return c.Next()
	}
}

func CurrentPrincipal(c fiber.Ctx) (Principal, bool) {
	principal, ok := c.Locals(principalContextKey{}).(Principal)
	return principal, ok
}

func unauthenticatedError() error {
	return &httperror.AppError{
		Code:       "unauthenticated",
		Message:    "Authentication required",
		HTTPStatus: fiber.StatusUnauthorized,
	}
}
