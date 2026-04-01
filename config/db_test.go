package config_test

import (
	"os"
	"testing"

	"github.com/weprodev/go-pkg/config"
)

func TestApplyDatabaseOverrides_NoEnv(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:         "original-host",
		Port:         5432,
		User:         "original-user",
		DBName:       "original-db",
		Password:     "original-pass",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}

	result := config.ApplyDatabaseOverrides(cfg)

	if result.Host != "original-host" {
		t.Errorf("Host = %q, want %q", result.Host, "original-host")
	}
	if result.Port != 5432 {
		t.Errorf("Port = %d, want 5432", result.Port)
	}
	if result.MaxOpenConns != 10 {
		t.Errorf("MaxOpenConns = %d, want 10", result.MaxOpenConns)
	}
}

func TestApplyDatabaseOverrides_DatabaseURL(t *testing.T) {
	const url = "postgres://user:pass@host:5432/dbname"
	t.Setenv("DATABASE_URL", url)
	t.Setenv("DB_MAX_OPEN_CONNS", "25")
	t.Setenv("DB_MAX_IDLE_CONNS", "12")

	result := config.ApplyDatabaseOverrides(config.DatabaseConfig{
		Host:         "fallback",
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	})

	if result.DatabaseURL != url {
		t.Errorf("DatabaseURL = %q, want %q", result.DatabaseURL, url)
	}
	if result.MaxOpenConns != 25 {
		t.Errorf("MaxOpenConns = %d, want 25", result.MaxOpenConns)
	}
	if result.MaxIdleConns != 12 {
		t.Errorf("MaxIdleConns = %d, want 12", result.MaxIdleConns)
	}
	// Individual fields should not be overwritten when DATABASE_URL is set.
	if result.Host != "fallback" {
		t.Errorf("Host = %q, want %q (should be unchanged)", result.Host, "fallback")
	}
}

func TestApplyDatabaseOverrides_IndividualFields(t *testing.T) {
	t.Setenv("DB_HOST", "env-host")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "env-user")
	t.Setenv("DB_NAME", "env-db")
	t.Setenv("DB_PASSWORD", "env-pass")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("DB_CONNECTION_NAME", "env-conn")
	t.Setenv("DB_MAX_OPEN_CONNS", "20")
	t.Setenv("DB_MAX_IDLE_CONNS", "10")

	result := config.ApplyDatabaseOverrides(config.DatabaseConfig{})

	checks := map[string]string{
		result.Host:           "env-host",
		result.User:           "env-user",
		result.DBName:         "env-db",
		result.Password:       "env-pass",
		result.SSLMode:        "require",
		result.ConnectionName: "env-conn",
	}
	for got, want := range checks {
		if got != want {
			t.Errorf("field = %q, want %q", got, want)
		}
	}
	if result.Port != 3306 {
		t.Errorf("Port = %d, want 3306", result.Port)
	}
	if result.MaxOpenConns != 20 {
		t.Errorf("MaxOpenConns = %d, want 20", result.MaxOpenConns)
	}
}

func TestApplyDatabaseOverrides_PartialEnv(t *testing.T) {
	t.Setenv("DB_HOST", "env-host")
	t.Setenv("DB_USER", "env-user")

	result := config.ApplyDatabaseOverrides(config.DatabaseConfig{
		Host:   "original-host",
		Port:   5432,
		User:   "original-user",
		DBName: "original-db",
	})

	if result.Host != "env-host" {
		t.Errorf("Host = %q, want %q", result.Host, "env-host")
	}
	// Non-overidden fields must be preserved.
	if result.DBName != "original-db" {
		t.Errorf("DBName = %q, want %q (should be unchanged)", result.DBName, "original-db")
	}
}

// ─── ValidateDatabaseConfig ───────────────────────────────────────────────────

func TestValidateDatabaseConfig_Valid(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.DatabaseConfig
	}{
		{
			name: "valid with host",
			cfg: config.DatabaseConfig{
				Host: "localhost", Port: 5432,
				User: "u", Password: "p", DBName: "db",
			},
		},
		{
			name: "valid with connection name",
			cfg: config.DatabaseConfig{
				ConnectionName: "proj:region:inst", Port: 5432,
				User: "u", Password: "p", DBName: "db",
			},
		},
		{
			name: "valid with DATABASE_URL",
			cfg:  config.DatabaseConfig{DatabaseURL: "postgres://u:p@host/db"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, _ := config.ValidateDatabaseConfig(tt.cfg)
			if len(errs) != 0 {
				t.Errorf("expected no errors, got: %v", errs)
			}
		})
	}
}

func TestValidateDatabaseConfig_Errors(t *testing.T) {
	tests := []struct {
		name         string
		cfg          config.DatabaseConfig
		wantErrCount int
	}{
		{
			name:         "missing host and connection name",
			cfg:          config.DatabaseConfig{Port: 5432, User: "u", Password: "p", DBName: "db"},
			wantErrCount: 1,
		},
		{
			name:         "invalid port — zero",
			cfg:          config.DatabaseConfig{Host: "h", Port: 0, User: "u", Password: "p", DBName: "db"},
			wantErrCount: 1,
		},
		{
			name:         "invalid port — too high",
			cfg:          config.DatabaseConfig{Host: "h", Port: 99999, User: "u", Password: "p", DBName: "db"},
			wantErrCount: 1,
		},
		{
			name:         "missing user",
			cfg:          config.DatabaseConfig{Host: "h", Port: 5432, Password: "p", DBName: "db"},
			wantErrCount: 1,
		},
		{
			name:         "missing password",
			cfg:          config.DatabaseConfig{Host: "h", Port: 5432, User: "u", DBName: "db"},
			wantErrCount: 1,
		},
		{
			name:         "missing dbname",
			cfg:          config.DatabaseConfig{Host: "h", Port: 5432, User: "u", Password: "p"},
			wantErrCount: 1,
		},
		{
			name:         "empty config — all errors",
			cfg:          config.DatabaseConfig{},
			wantErrCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, _ := config.ValidateDatabaseConfig(tt.cfg)
			if len(errs) != tt.wantErrCount {
				t.Errorf("error count = %d, want %d: %v", len(errs), tt.wantErrCount, errs)
			}
		})
	}
}

func TestValidateDatabaseConfig_WithRealEnv(t *testing.T) {
	_ = os.Setenv("DB_HOST", "real-host")
	_ = os.Setenv("DB_USER", "real-user")
	_ = os.Setenv("DB_PASSWORD", "real-pass")
	_ = os.Setenv("DB_NAME", "real-db")

	defer func() {
		_ = os.Unsetenv("DB_HOST")
		_ = os.Unsetenv("DB_USER")
		_ = os.Unsetenv("DB_PASSWORD")
		_ = os.Unsetenv("DB_NAME")
	}()

	cfg := config.DatabaseConfig{Host: "real-host", Port: 5432, User: "real-user", Password: "real-pass", DBName: "real-db"}
	errs, warns := config.ValidateDatabaseConfig(cfg)

	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
	if len(warns) != 0 {
		t.Errorf("expected no warnings, got: %v", warns)
	}
}
