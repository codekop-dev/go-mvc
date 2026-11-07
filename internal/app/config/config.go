package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv   string
	Name     string
	Version  string
	Timezone string
	HTTP     HTTPConfig
	Database DatabaseConfig
}

type HTTPConfig struct {
	Address      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	Socket   string
	Name     string
	User     string
	Password string
	SSLMode  string
}

// Load reads application settings from environment variables and applies safe defaults.
func Load() Config {
	return Config{
		AppEnv:   env("APP_ENV", "production"),
		Name:     env("APP_NAME", "Codekop Go MVC"),
		Version:  env("APP_VERSION", "1.0.0"),
		Timezone: env("APP_TIMEZONE", "Asia/Jakarta"),
		HTTP: HTTPConfig{
			Address:      env("HTTP_ADDR", ":8080"),
			ReadTimeout:  duration("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: duration("HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  duration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Driver:   normalizeDriver(env("DB_DRIVER", "mysql")),
			Host:     env("DB_HOST", "127.0.0.1"),
			Port:     integer("DB_PORT", 3306),
			Socket:   os.Getenv("DB_SOCKET"),
			Name:     os.Getenv("DB_NAME"),
			User:     env("DB_USER", "root"),
			Password: os.Getenv("DB_PASSWORD"),
			SSLMode:  env("DB_SSLMODE", "disable"),
		},
	}
}

// IsDevelopment reports whether debug behavior should be enabled.
func (c Config) IsDevelopment() bool {
	return c.AppEnv == "development" || c.AppEnv == "dev" || c.AppEnv == "testing"
}

// Location resolves the configured timezone and falls back to UTC if invalid.
func (c Config) Location() *time.Location {
	location, err := time.LoadLocation(c.Timezone)
	if err != nil {
		return time.UTC
	}
	return location
}

// LoadDotEnv loads a small .env file without replacing variables supplied by the shell.
// It intentionally stays dependency-free so a PHP developer can see the whole bootstrap.
// LoadDotEnv loads simple KEY=VALUE lines without overriding shell variables.
func LoadDotEnv(path string) error {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return fmt.Errorf("invalid .env line %d", lineNumber)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

// env returns a non-empty environment value or its fallback.
func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

// integer parses a positive integer environment setting.
func integer(key string, fallback int) int {
	value, err := strconv.Atoi(env(key, strconv.Itoa(fallback)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

// duration parses a positive Go duration environment setting.
func duration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(env(key, fallback.String()))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

// normalizeDriver maps common PostgreSQL aliases and defaults to MySQL.
func normalizeDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgres", "postgresql", "pg":
		return "postgres"
	default:
		return "mysql"
	}
}
