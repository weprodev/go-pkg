package config

import (
	"fmt"
	"time"
)

// DatabaseConfig holds all parameters needed to open a database connection.
type DatabaseConfig struct {
	// Host is the hostname or IP of the database server (TCP connections).
	Host string `yaml:"host"`
	// ConnectionName is the managed-database instance identifier used with
	// Unix-socket connections (e.g. GCP Cloud SQL, AWS RDS Proxy socket path).
	ConnectionName string        `yaml:"connection_name"`
	DatabaseURL    string        `yaml:"url"`
	Port           int           `yaml:"port"`
	User           string        `yaml:"user"`
	Password       string        `yaml:"password"`
	DBName         string        `yaml:"dbname"`
	SSLMode        string        `yaml:"sslmode"`
	MaxOpenConns   int           `yaml:"max_open_conns"`
	MaxIdleConns   int           `yaml:"max_idle_conns"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
}

// ApplyDatabaseOverrides overlays standard DB_* environment variables onto
// the provided DatabaseConfig and returns the updated value.
//
// DATABASE_URL takes precedence over individual fields: when it is set,
// individual host/user/password fields are left unchanged and the URL is
// used directly by the driver.
func ApplyDatabaseOverrides(config DatabaseConfig) DatabaseConfig {
	config.MaxOpenConns = GetEnvInt("DB_MAX_OPEN_CONNS", config.MaxOpenConns)
	config.MaxIdleConns = GetEnvInt("DB_MAX_IDLE_CONNS", config.MaxIdleConns)

	if url := GetEnv("DATABASE_URL", ""); url != "" {
		config.DatabaseURL = url
		return config
	}

	config.Host = GetEnv("DB_HOST", config.Host)
	config.User = GetEnv("DB_USER", config.User)
	config.DBName = GetEnv("DB_NAME", config.DBName)
	config.Port = GetEnvInt("DB_PORT", config.Port)
	config.SSLMode = GetEnv("DB_SSLMODE", config.SSLMode)
	config.Password = GetEnv("DB_PASSWORD", config.Password)
	config.ConnectionName = GetEnv("DB_CONNECTION_NAME", config.ConnectionName)

	return config
}

// ValidateDatabaseConfig validates a DatabaseConfig and returns two slices:
// errors (blocking) and warnings (advisory). An empty errors slice means the
// configuration is usable.
func ValidateDatabaseConfig(config DatabaseConfig) (errs []string, warnings []string) {
	// DATABASE_URL satisfies all connection requirements on its own.
	if config.DatabaseURL != "" {
		return nil, nil
	}

	if config.Host == "" && config.ConnectionName == "" {
		errs = append(errs, "either DB_HOST or DB_CONNECTION_NAME is required")
	}

	if config.Port <= 0 || config.Port > 65535 {
		errs = append(errs, fmt.Sprintf("DB_PORT must be 1–65535, got %d", config.Port))
	}

	if config.User == "" {
		errs = append(errs, "DB_USER is required")
	}

	if config.Password == "" {
		errs = append(errs, "DB_PASSWORD is required")
	}

	if config.DBName == "" {
		errs = append(errs, "DB_NAME is required")
	}

	return errs, warnings
}
