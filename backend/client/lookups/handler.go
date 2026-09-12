package lookups

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thienty1207/BWP-Staff/backend/client/auth"
	"github.com/thienty1207/BWP-Staff/backend/shared/httperror"
)

var errLookupServiceNotConfigured = &httperror.AppError{
	Code:       "internal_server_error",
	Message:    "Internal server error",
	HTTPStatus: fiber.StatusInternalServerError,
}

func RegisterRoutes(api fiber.Router, pool *pgxpool.Pool, authService *auth.Service) {
	handler := &handler{service: NewService(NewRepository(pool))}
	api.Get("/departments", authService.RequireAuth(), handler.departments)
	api.Get("/locations", authService.RequireAuth(), handler.locations)
}

type handler struct {
	service *Service
}

func (handler *handler) departments(c fiber.Ctx) error {
	departments, err := handler.service.ListDepartments(c.Context())
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(DepartmentsResponse{Departments: departments})
}

func (handler *handler) locations(c fiber.Ctx) error {
	locations, err := handler.service.ListLocations(c.Context())
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(LocationsResponse{Locations: locations})
}
