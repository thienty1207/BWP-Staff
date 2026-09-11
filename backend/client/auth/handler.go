package auth

import (
	"encoding/json"
	"errors"
	"net"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thienty1207/BWP-Staff/backend/config"
	"github.com/thienty1207/BWP-Staff/backend/shared/httperror"
)

func RegisterRoutes(api fiber.Router, pool *pgxpool.Pool, settings config.Config) *Service {
	repository := NewRepository(pool)
	service := NewService(repository, time.Duration(settings.AuthSessionTTLHours)*time.Hour)
	handler := &handler{
		service:      service,
		secureCookie: !strings.EqualFold(settings.AppEnv, "development"),
	}

	authRoutes := api.Group("/auth")
	authRoutes.Post("/login", handler.login)
	authRoutes.Post("/logout", handler.logout)
	authRoutes.Get("/me", service.RequireAuth(), handler.me)
	return service
}

type handler struct {
	service      *Service
	secureCookie bool
}

func (handler *handler) login(c fiber.Ctx) error {
	var input loginRequest
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return invalidRequestError()
	}
	input, err := validateLoginRequest(input)
	if err != nil {
		return invalidRequestError()
	}

	result, err := handler.service.Login(c.Context(), input, requestMetadata(c))
	if errors.Is(err, ErrInvalidCredentials) {
		return invalidCredentialsError()
	}
	if err != nil {
		return err
	}

	c.Cookie(sessionCookie(result.RawToken, result.ExpiresAt, handler.secureCookie))
	return c.Status(fiber.StatusOK).JSON(struct {
		User UserIdentity `json:"user"`
	}{User: result.User})
}

func (handler *handler) logout(c fiber.Ctx) error {
	rawToken := c.Cookies(sessionCookieName)
	err := handler.service.Logout(c.Context(), rawToken)
	if err != nil {
		return err
	}
	c.Cookie(expiredSessionCookie(handler.secureCookie))
	return c.SendStatus(fiber.StatusNoContent)
}

func (handler *handler) me(c fiber.Ctx) error {
	principal, ok := CurrentPrincipal(c)
	if !ok {
		return unauthenticatedError()
	}
	return c.Status(fiber.StatusOK).JSON(struct {
		User UserIdentity `json:"user"`
	}{User: principal.User})
}

func validateLoginRequest(input loginRequest) (loginRequest, error) {
	input.Username = strings.TrimSpace(input.Username)
	if input.Username == "" || utf8.RuneCountInString(input.Username) > maxUsernameCharacters {
		return loginRequest{}, errors.New("invalid username")
	}
	if input.Password == "" || len(input.Password) > maxPasswordBytes {
		return loginRequest{}, errors.New("invalid password")
	}
	return input, nil
}

func requestMetadata(c fiber.Ctx) sessionMetadata {
	metadata := sessionMetadata{}
	if userAgent := capUTF8(c.Get(fiber.HeaderUserAgent), maxUserAgentBytes); userAgent != "" {
		metadata.UserAgent = &userAgent
	}
	if ip := net.ParseIP(c.IP()); ip != nil {
		address := ip.String()
		metadata.IPAddress = &address
	}
	return metadata
}

func capUTF8(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	value = value[:maxBytes]
	for len(value) > 0 && !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}

func sessionCookie(rawToken string, expiresAt time.Time, secure bool) *fiber.Cookie {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 1 {
		maxAge = 1
	}
	return &fiber.Cookie{
		Name:     sessionCookieName,
		Value:    rawToken,
		Expires:  expiresAt,
		MaxAge:   maxAge,
		Path:     "/",
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   secure,
		HTTPOnly: true,
	}
}

func expiredSessionCookie(secure bool) *fiber.Cookie {
	return &fiber.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Expires:  time.Unix(1, 0).UTC(),
		MaxAge:   -1,
		Path:     "/",
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   secure,
		HTTPOnly: true,
	}
}

func invalidRequestError() error {
	return &httperror.AppError{
		Code:       "invalid_request",
		Message:    "Invalid request",
		HTTPStatus: fiber.StatusBadRequest,
	}
}

func invalidCredentialsError() error {
	return &httperror.AppError{
		Code:       "invalid_credentials",
		Message:    "Invalid username or password",
		HTTPStatus: fiber.StatusUnauthorized,
	}
}
