package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/thienty1207/BWP-Staff/backend/config"
)

type statusResponse struct {
	Status string `json:"status"`
}

type errorResponse struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
	} `json:"error"`
}

func testAppSettings() config.Config {
	return config.Config{
		AppEnv:                        "development",
		FrontendOrigin:                "http://localhost:5173",
		DatabaseAcquireTimeoutSeconds: 1,
		BackendShutdownTimeoutSeconds: 1,
	}
}

func TestHealthEndpointReturnsJSONAndRequestID(t *testing.T) {
	response, err := New(AppState{}, testAppSettings()).Test(httptest.NewRequest(http.MethodGet, "/health", nil))
	if err != nil {
		t.Fatalf("request health endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}
	if contentType := response.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("expected JSON content type, got %q", contentType)
	}
	if response.Header.Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID response header")
	}

	var body statusResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("expected health status %q, got %q", "ok", body.Status)
	}
}

func TestHealthEndpointDoesNotRequireDatabase(t *testing.T) {
	server := New(AppState{}, testAppSettings())
	response, err := server.Test(httptest.NewRequest(http.MethodGet, "/health", nil))
	if err != nil {
		t.Fatalf("request health endpoint without database: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected health to remain available without database, got %d", response.StatusCode)
	}
}

func TestReadyEndpointReturnsSafeErrorWhenDatabaseIsUnavailable(t *testing.T) {
	pool := openAppTestPool(t)
	pool.Close()

	const requestID = "ready-error-request"
	request := httptest.NewRequest(http.MethodGet, "/ready", nil)
	request.Header.Set("X-Request-ID", requestID)
	response, err := New(AppState{DB: pool}, testAppSettings()).Test(request)
	if err != nil {
		t.Fatalf("request readiness endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", response.StatusCode)
	}
	if contentType := response.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("expected JSON content type, got %q", contentType)
	}
	if response.Header.Get("X-Request-ID") != requestID {
		t.Fatalf("expected request ID %q, got %q", requestID, response.Header.Get("X-Request-ID"))
	}

	var body errorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode readiness error: %v", err)
	}
	if body.Error.Code != "not_ready" || body.Error.Message != "Service not ready" {
		t.Fatalf("unexpected readiness error: %+v", body.Error)
	}
	if body.Error.RequestID != requestID {
		t.Fatalf("expected error request ID %q, got %q", requestID, body.Error.RequestID)
	}
	if strings.Contains(body.Error.Message, "127.0.0.1") || strings.Contains(body.Error.Message, "password") {
		t.Fatal("readiness error exposed database details")
	}
}

func TestReadyEndpointReturnsReadyForHealthyPostgreSQL(t *testing.T) {
	pool := openAppTestPool(t)

	response, err := New(AppState{DB: pool}, testAppSettings()).Test(httptest.NewRequest(http.MethodGet, "/ready", nil))
	if err != nil {
		t.Fatalf("request readiness endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}
	var body statusResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode readiness response: %v", err)
	}
	if body.Status != "ready" {
		t.Fatalf("expected readiness status %q, got %q", "ready", body.Status)
	}
}

func TestUnknownRouteReturnsStableJSONError(t *testing.T) {
	const requestID = "not-found-request"
	request := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	request.Header.Set("X-Request-ID", requestID)
	response, err := New(AppState{}, testAppSettings()).Test(request)
	if err != nil {
		t.Fatalf("request unknown route: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.StatusCode)
	}
	var body errorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode not-found response: %v", err)
	}
	if body.Error.Code != "not_found" || body.Error.Message != "Resource not found" {
		t.Fatalf("unexpected not-found error: %+v", body.Error)
	}
	if body.Error.RequestID != requestID {
		t.Fatalf("expected error request ID %q, got %q", requestID, body.Error.RequestID)
	}
}

func TestMethodNotAllowedReturnsStableJSONError(t *testing.T) {
	response, err := New(AppState{}, testAppSettings()).Test(httptest.NewRequest(http.MethodPost, "/health", nil))
	if err != nil {
		t.Fatalf("request unsupported method: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", response.StatusCode)
	}
	var body errorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode method-not-allowed response: %v", err)
	}
	if body.Error.Code != "method_not_allowed" || body.Error.Message != "Method not allowed" {
		t.Fatalf("unexpected method-not-allowed error: %+v", body.Error)
	}
}

func TestKnownAppErrorUsesSafeJSONError(t *testing.T) {
	server := New(AppState{}, testAppSettings())
	server.Get("/test-known-error", func(fiber.Ctx) error {
		return &AppError{
			Code:       "example_error",
			Message:    "Example failure",
			HTTPStatus: http.StatusTeapot,
		}
	})

	response, err := server.Test(httptest.NewRequest(http.MethodGet, "/test-known-error", nil))
	if err != nil {
		t.Fatalf("request known application error: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusTeapot {
		t.Fatalf("expected status 418, got %d", response.StatusCode)
	}
	var body errorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode known application error: %v", err)
	}
	if body.Error.Code != "example_error" || body.Error.Message != "Example failure" {
		t.Fatalf("unexpected known application error: %+v", body.Error)
	}
	if body.Error.RequestID == "" {
		t.Fatal("expected known application error request ID")
	}
}

func TestUnknownHandlerErrorIsSafeAndLogged(t *testing.T) {
	const (
		internalMessage = "database password must not be exposed"
		requestID       = "internal-error-request"
	)
	server := New(AppState{}, testAppSettings())
	server.Get("/test-internal-error", func(fiber.Ctx) error {
		return errors.New(internalMessage)
	})

	var logs bytes.Buffer
	previousWriter := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(previousWriter)

	request := httptest.NewRequest(http.MethodGet, "/test-internal-error", nil)
	request.Header.Set("X-Request-ID", requestID)
	response, err := server.Test(request)
	if err != nil {
		t.Fatalf("request internal error: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read internal error response: %v", err)
	}
	if strings.Contains(string(body), internalMessage) {
		t.Fatal("internal error text leaked into response")
	}
	if !strings.Contains(logs.String(), internalMessage) || !strings.Contains(logs.String(), "request_id="+requestID) {
		t.Fatalf("internal error log lacks error/request context: %q", logs.String())
	}
}

func TestPanicRecoveryReturnsSafeErrorAndKeepsAppAlive(t *testing.T) {
	const panicMessage = "panic secret"
	server := New(AppState{}, testAppSettings())
	server.Get("/test-panic", func(fiber.Ctx) error {
		panic(panicMessage)
	})

	response, err := server.Test(httptest.NewRequest(http.MethodGet, "/test-panic", nil))
	if err != nil {
		t.Fatalf("request panic route: %v", err)
	}
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatalf("read panic response: %v", err)
	}
	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.StatusCode)
	}
	if strings.Contains(string(body), panicMessage) {
		t.Fatal("panic detail leaked into response")
	}

	healthResponse, err := server.Test(httptest.NewRequest(http.MethodGet, "/health", nil))
	if err != nil {
		t.Fatalf("request health after panic: %v", err)
	}
	defer healthResponse.Body.Close()
	if healthResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected app to remain alive after panic, got status %d", healthResponse.StatusCode)
	}
}

func TestConfiguredOriginReceivesCredentialedCORSHeaders(t *testing.T) {
	const origin = "http://localhost:5173"
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set("Origin", origin)
	response, err := New(AppState{}, testAppSettings()).Test(request)
	if err != nil {
		t.Fatalf("request configured CORS origin: %v", err)
	}
	defer response.Body.Close()

	if response.Header.Get("Access-Control-Allow-Origin") != origin {
		t.Fatalf("expected CORS origin %q, got %q", origin, response.Header.Get("Access-Control-Allow-Origin"))
	}
	if response.Header.Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("expected credentialed CORS response")
	}
}

func TestConfiguredOriginPreflightReceivesExplicitCORSPolicy(t *testing.T) {
	const origin = "http://localhost:5173"
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/future", nil)
	request.Header.Set("Origin", origin)
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "Content-Type")
	response, err := New(AppState{}, testAppSettings()).Test(request)
	if err != nil {
		t.Fatalf("request configured CORS preflight: %v", err)
	}
	defer response.Body.Close()

	if response.Header.Get("Access-Control-Allow-Origin") != origin {
		t.Fatalf("expected CORS preflight origin %q, got %q", origin, response.Header.Get("Access-Control-Allow-Origin"))
	}
	if response.Header.Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("expected credentialed CORS preflight response")
	}
	if !strings.Contains(response.Header.Get("Access-Control-Allow-Methods"), http.MethodPost) {
		t.Fatalf("expected POST in CORS preflight methods, got %q", response.Header.Get("Access-Control-Allow-Methods"))
	}
}

func TestUnrelatedOriginDoesNotReceivePermissiveCORSHeaders(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set("Origin", "https://unrelated.example.com")
	response, err := New(AppState{}, testAppSettings()).Test(request)
	if err != nil {
		t.Fatalf("request unrelated CORS origin: %v", err)
	}
	defer response.Body.Close()

	allowOrigin := response.Header.Get("Access-Control-Allow-Origin")
	if allowOrigin == "*" || allowOrigin != "" {
		t.Fatalf("unrelated origin received permissive CORS header: %q", allowOrigin)
	}
}

func TestRequestLoggingIncludesRequestContext(t *testing.T) {
	const requestID = "logging-request"
	var logs bytes.Buffer
	previousWriter := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(previousWriter)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set("X-Request-ID", requestID)
	response, err := New(AppState{}, testAppSettings()).Test(request)
	if err != nil {
		t.Fatalf("request logged health endpoint: %v", err)
	}
	response.Body.Close()

	logged := logs.String()
	for _, field := range []string{
		"request_id=" + requestID,
		"method=GET",
		"path=/health",
		"status=200",
		"latency=",
	} {
		if !strings.Contains(logged, field) {
			t.Fatalf("request log missing %q: %q", field, logged)
		}
	}
}
