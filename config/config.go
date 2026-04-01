// Package config provides generic configuration loading utilities for Go services.
//
// It defines common configuration structures (server, logging, security, database)
// and helpers to load them from YAML or JSON files with environment-variable
// overrides. The package intentionally contains no application-specific fields;
// consuming services embed or compose these types to extend them.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the base configuration structure for a generic web service.
// Services may embed this type and add their own domain-specific fields.
type Config struct {
	Environment Environment    `yaml:"environment"`
	Server      ServerConfig   `yaml:"server"`
	Logging     LoggingConfig  `yaml:"logging"`
	Security    SecurityConfig `yaml:"security"`
	DB          DatabaseConfig `yaml:"db"`
}

// ServerConfig holds HTTP server tuning parameters.
type ServerConfig struct {
	Port              int           `yaml:"port"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
}

// LoggingConfig controls log level, format, and output destination.
type LoggingConfig struct {
	Level      string `yaml:"level"`       // debug | info | warn | error
	Format     string `yaml:"format"`      // json | text
	OutputPath string `yaml:"output_path"` // file path or "stdout"
}

// SecurityConfig holds generic security settings.
// JWT, CORS, and token-expiry are common to virtually every authenticated web service.
type SecurityConfig struct {
	JWTSecret          string        `yaml:"jwt_secret"`
	TokenExpiry        time.Duration `yaml:"token_expiry"`
	RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry"`
	RefreshTokenPath   string        `yaml:"refresh_token_path"`
	AllowedHosts       []string      `yaml:"allowed_hosts"`
	CORS               CORSConfig    `yaml:"cors"`
}

// CORSConfig configures Cross-Origin Resource Sharing.
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}

// LoadConfig loads configuration from a YAML file and applies
// environment-variable overrides on top.
//
// If configPath is empty, FindConfigFile is used to locate a config file
// based on the ENVIRONMENT variable.
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = FindConfigFile()
	}

	config, err := loadConfigFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config %s: %w", configPath, err)
	}

	ApplyEnvironmentOverrides(config)

	if errs := ValidateConfig(config); len(errs) > 0 {
		return nil, fmt.Errorf("config validation failed: %s", strings.Join(errs, "; "))
	}

	return config, nil
}

// loadConfigFromFile reads and unmarshals a YAML config file.
func loadConfigFromFile(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}

	return &config, nil
}

// ApplyEnvironmentOverrides overlays environment variables onto an existing
// Config. It is exported so that services with extended config structs can
// call it after unmarshalling their own YAML.
func ApplyEnvironmentOverrides(config *Config) {
	config.Environment = Environment(GetEnv("ENVIRONMENT", config.Environment.String()))
	config.DB = ApplyDatabaseOverrides(config.DB)

	// Security
	config.Security.JWTSecret = GetEnv("JWT_SECRET", config.Security.JWTSecret)
	if v := GetEnv("TOKEN_EXPIRY", ""); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.Security.TokenExpiry = d
		}
	}
	if v := GetEnv("REFRESH_TOKEN_EXPIRY", ""); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.Security.RefreshTokenExpiry = d
		}
	}

	// CORS
	if origins := GetEnv("CORS_ALLOWED_ORIGINS", ""); origins != "" {
		config.Security.CORS.AllowedOrigins = strings.Split(origins, ",")
	}

	// Server
	config.Server.Port = GetEnvInt("SERVER_PORT", config.Server.Port)
	if v := GetEnv("SERVER_READ_TIMEOUT", ""); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.Server.ReadTimeout = d
		}
	}
	if v := GetEnv("SERVER_WRITE_TIMEOUT", ""); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.Server.WriteTimeout = d
		}
	}
	if v := GetEnv("SERVER_IDLE_TIMEOUT", ""); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.Server.IdleTimeout = d
		}
	}
	if v := GetEnv("SERVER_READ_HEADER_TIMEOUT", ""); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.Server.ReadHeaderTimeout = d
		}
	}

	// Logging
	config.Logging.Level = GetEnv("LOG_LEVEL", config.Logging.Level)
	config.Logging.Format = GetEnv("LOG_FORMAT", config.Logging.Format)
	config.Logging.OutputPath = GetEnv("LOG_OUTPUT_PATH", config.Logging.OutputPath)
}

// ValidateConfig checks required fields and returns a slice of error messages.
// An empty slice means the config is valid.
func ValidateConfig(config *Config) []string {
	var errs []string
	dbErrs, _ := ValidateDatabaseConfig(config.DB)
	errs = append(errs, dbErrs...)
	return errs
}

// IsSensitiveKey reports whether an environment-variable name likely holds
// sensitive data (passwords, tokens, API keys, secrets).
func IsSensitiveKey(name string) bool {
	upper := strings.ToUpper(name)
	for _, kw := range []string{"PASSWORD", "SECRET", "TOKEN", "API_KEY", "PRIVATE_KEY"} {
		if strings.Contains(upper, kw) {
			return true
		}
	}
	return false
}

// MaskSecret masks the value for safe logging. The first 4 characters are
// revealed so operators can identify which secret is set without exposing it.
func MaskSecret(value string) string {
	if value == "" {
		return "[NOT SET]"
	}
	if len(value) < 8 {
		return "[SET]"
	}
	return "[SET - " + value[:4] + "****]"
}

// FindConfigFile returns the path of the first config file found, searching
// a standard set of locations ordered by precedence.
func FindConfigFile() string {
	environment := GetEnv("ENVIRONMENT", Local.String())

	locations := []string{
		fmt.Sprintf("configs/%s.yml", environment),
		"configs/local.yml",
		"configs/config.yml",
		"config.yml",
		"local.yml",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return fmt.Sprintf("configs/%s.yml", environment)
}
