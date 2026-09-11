package app

import (
	"errors"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// AppError is a safe, explicit error that a handler may return to its client.
type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (err *AppError) Error() string {
	if err == nil {
		return ""
	}
	return err.Message
}

type httpErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

type httpErrorResponse struct {
	Error httpErrorBody `json:"error"`
}

func handleError(c fiber.Ctx, err error) error {
	requestID := requestid.FromContext(c)

	var appErr *AppError
	if errors.As(err, &appErr) {
		return writeError(c, appErr.HTTPStatus, appErr.Code, appErr.Message, requestID)
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		status, code, message := safeFiberError(fiberErr.Code)
		return writeError(c, status, code, message, requestID)
	}

	log.Printf("request_id=%s method=%s path=%s internal_error=%v", requestID, c.Method(), c.Path(), err)
	return writeError(c, http.StatusInternalServerError, "internal_server_error", "Internal server error", requestID)
}

func writeError(c fiber.Ctx, status int, code, message, requestID string) error {
	return c.Status(status).JSON(httpErrorResponse{
		Error: httpErrorBody{
			Code:      code,
			Message:   message,
			RequestID: requestID,
		},
	})
}

func safeFiberError(status int) (int, string, string) {
	switch status {
	case http.StatusNotFound:
		return status, "not_found", "Resource not found"
	case http.StatusMethodNotAllowed:
		return status, "method_not_allowed", "Method not allowed"
	case http.StatusBadRequest:
		return status, "bad_request", "Bad request"
	case http.StatusUnauthorized:
		return status, "unauthorized", "Unauthorized"
	case http.StatusForbidden:
		return status, "forbidden", "Forbidden"
	case http.StatusRequestEntityTooLarge:
		return status, "request_too_large", "Request too large"
	case http.StatusServiceUnavailable:
		return status, "service_unavailable", "Service unavailable"
	default:
		if status < http.StatusBadRequest || status > 599 {
			return http.StatusInternalServerError, "internal_server_error", "Internal server error"
		}
		return status, "http_error", "HTTP request failed"
	}
}

func statusForError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	var appErr *AppError
	if errors.As(err, &appErr) && appErr.HTTPStatus >= 400 && appErr.HTTPStatus <= 599 {
		return appErr.HTTPStatus
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) && fiberErr.Code >= 400 && fiberErr.Code <= 599 {
		return fiberErr.Code
	}

	return http.StatusInternalServerError
}
