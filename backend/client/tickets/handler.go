package tickets

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thienty1207/BWP-Staff/backend/client/auth"
	"github.com/thienty1207/BWP-Staff/backend/shared/httperror"
)

var errTicketServiceNotConfigured = &httperror.AppError{
	Code:       "internal_server_error",
	Message:    "Internal server error",
	HTTPStatus: fiber.StatusInternalServerError,
}

func RegisterRoutes(api fiber.Router, pool *pgxpool.Pool, authService *auth.Service) {
	handler := &handler{service: NewService(NewRepository(pool))}
	api.Get("/tickets", authService.RequireAuth(), handler.list)
}

type handler struct {
	service *Service
}

func (handler *handler) list(c fiber.Ctx) error {
	query, err := parseListQueryValues(c.Queries())
	if err != nil {
		return &httperror.AppError{
			Code:       "invalid_request",
			Message:    "Invalid request",
			HTTPStatus: fiber.StatusBadRequest,
		}
	}
	response, err := handler.service.List(c.Context(), query)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(response)
}
