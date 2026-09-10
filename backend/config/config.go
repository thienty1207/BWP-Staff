package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	defaultMaxConnections        int32 = 10
	defaultMinConnections        int32 = 1
	maxAllowedConnections        int32 = 100
	defaultAcquireTimeoutSeconds       = 5
	maxAcquireTimeoutSeconds           = 60
)

type Config struct {
	DatabaseURL                   string
	DatabaseMaxConnections        int32
	DatabaseMinConnections        int32
	DatabaseAcquireTimeoutSeconds int
	AppEnv                        string
	SeedDevelopmentData           bool
	SeedAdmin                     SeedAdmin
}

type SeedAdmin struct {
	Username       string
	Password       string
	EmployeeCode   string
	FullName       string
	DepartmentCode string
}

func Load() (Config, error) {
	databaseURL, err := required("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	parsedURL, err := url.Parse(databaseURL)
	if err != nil || (parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql") || parsedURL.Host == "" {
		return Config{}, fmt.Errorf("invalid DATABASE_URL")
	}
	if err := validateDatabaseTLS(parsedURL); err != nil {
		return Config{}, err
	}

	maxConnections, err := int32Value("DATABASE_MAX_CONNECTIONS", defaultMaxConnections)
	if err != nil {
		return Config{}, err
	}
	minConnections, err := int32Value("DATABASE_MIN_CONNECTIONS", defaultMinConnections)
	if err != nil {
		return Config{}, err
	}
	if maxConnections <= 0 || maxConnections > maxAllowedConnections || minConnections < 0 || minConnections > maxConnections {
		return Config{}, fmt.Errorf("invalid database pool bounds")
	}

	acquireTimeoutSeconds, err := intValue("DATABASE_ACQUIRE_TIMEOUT_SECONDS", defaultAcquireTimeoutSeconds)
	if err != nil {
		return Config{}, err
	}
	if acquireTimeoutSeconds <= 0 || acquireTimeoutSeconds > maxAcquireTimeoutSeconds {
		return Config{}, fmt.Errorf("DATABASE_ACQUIRE_TIMEOUT_SECONDS must be between 1 and %d", maxAcquireTimeoutSeconds)
	}

	appEnv := valueOrDefault("APP_ENV", "development")
	seedDevelopmentData, err := boolValue("SEED_DEVELOPMENT_DATA", false)
	if err != nil {
		return Config{}, err
	}

	config := Config{
		DatabaseURL:                   databaseURL,
		DatabaseMaxConnections:        maxConnections,
		DatabaseMinConnections:        minConnections,
		DatabaseAcquireTimeoutSeconds: acquireTimeoutSeconds,
		AppEnv:                        appEnv,
		SeedDevelopmentData:           seedDevelopmentData,
	}
	if seedDevelopmentData && strings.EqualFold(appEnv, "development") {
		config.SeedAdmin, err = seedAdminFromEnvironment()
		if err != nil {
			return Config{}, err
		}
	}

	return config, nil
}

func seedAdminFromEnvironment() (SeedAdmin, error) {
	password, err := required("SEED_ADMIN_PASSWORD")
	if err != nil {
		return SeedAdmin{}, err
	}
	return SeedAdmin{
		Username:       valueOrDefault("SEED_ADMIN_USERNAME", "hothienty"),
		Password:       password,
		EmployeeCode:   valueOrDefault("SEED_ADMIN_EMPLOYEE_CODE", "IT-ADMIN-001"),
		FullName:       valueOrDefault("SEED_ADMIN_FULL_NAME", "Ho Thien Ty"),
		DepartmentCode: valueOrDefault("SEED_ADMIN_DEPARTMENT_CODE", "IT"),
	}, nil
}

func required(name string) (string, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", fmt.Errorf("missing %s", name)
	}
	return value, nil
}

func valueOrDefault(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func int32Value(name string, fallback int32) (int32, error) {
	value := valueOrDefault(name, strconv.FormatInt(int64(fallback), 10))
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	return int32(parsed), nil
}

func intValue(name string, fallback int) (int, error) {
	value := valueOrDefault(name, strconv.Itoa(fallback))
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	return parsed, nil
}

func boolValue(name string, fallback bool) (bool, error) {
	value := valueOrDefault(name, strconv.FormatBool(fallback))
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid %s", name)
	}
}

func validateDatabaseTLS(databaseURL *url.URL) error {
	host := strings.TrimSuffix(strings.ToLower(databaseURL.Hostname()), ".")
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return nil
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return nil
	}

	switch strings.ToLower(strings.TrimSpace(databaseURL.Query().Get("sslmode"))) {
	case "require", "verify-ca", "verify-full":
		return nil
	default:
		return fmt.Errorf("DATABASE_URL for a non-loopback database must set sslmode=require, verify-ca, or verify-full")
	}
}
