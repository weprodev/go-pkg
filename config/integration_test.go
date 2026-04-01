package config_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/weprodev/go-pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigIntegration tests the entire config loading and validation process
func TestConfigIntegration(t *testing.T) {
	tests := []struct {
		name           string
		configFile     string
		envVars        map[string]string
		expectedEnv    config.Environment
		shouldFail     bool
		expectedErrors []string
	}{
		{
			name:       "local config with all values set",
			configFile: "test-data/local.yml",
			envVars: map[string]string{
				"ENVIRONMENT": "local",
			},
			expectedEnv: config.Local,
			shouldFail:  false,
		},
		{
			name:       "staging config with environment overrides",
			configFile: "test-data/staging.yml",
			envVars: map[string]string{
				"ENVIRONMENT": "staging",
				"DB_HOST":     "staging-db.example.com",
				"DB_USER":     "staging_user",
				"DB_PASSWORD": "staging_password",
				"DB_NAME":     "staging_db",
				"LOG_LEVEL":   "info",
			},
			expectedEnv: config.Staging,
			shouldFail:  false,
		},
		{
			name:       "production config with environment overrides",
			configFile: "test-data/production.yml",
			envVars: map[string]string{
				"ENVIRONMENT": "production",
				"DB_HOST":     "prod-db.example.com",
				"DB_USER":     "prod_user",
				"DB_PASSWORD": "prod_password",
				"DB_NAME":     "prod_db",
				"LOG_LEVEL":   "warn",
			},
			expectedEnv: config.Production,
			shouldFail:  false,
		},
		{
			name:       "staging config missing required environment variables",
			configFile: "test-data/staging.yml",
			envVars: map[string]string{
				"ENVIRONMENT": "staging",
			},
			expectedEnv: config.Staging,
			shouldFail:  true,
		},
		{
			name:       "production config with DATABASE_URL",
			configFile: "test-data/production.yml",
			envVars: map[string]string{
				"ENVIRONMENT":  "production",
				"DATABASE_URL": "postgres://user:pass@host:5432/dbname?sslmode=require",
			},
			expectedEnv: config.Production,
			shouldFail:  false, // This should succeed because DATABASE_URL is provided
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			cfg, err := config.LoadConfig(tt.configFile)

			if tt.shouldFail {
				require.Error(t, err, "Expected configuration to fail but it succeeded")
				return
			}

			require.NoError(t, err, "Configuration should load successfully")
			require.NotNil(t, cfg, "Configuration should not be nil")

			assert.Equal(t, tt.expectedEnv, cfg.Environment, "Environment should match expected value")
			assert.Greater(t, cfg.Server.Port, 0, "Server port should be positive")
			assert.Greater(t, cfg.Server.ReadTimeout, time.Duration(0), "Read timeout should be positive")
			assert.Greater(t, cfg.Server.WriteTimeout, time.Duration(0), "Write timeout should be positive")

			assert.NotEmpty(t, cfg.Logging.Level, "Log level should not be empty")
			assert.NotEmpty(t, cfg.Logging.Format, "Log format should not be empty")
			assert.NotEmpty(t, cfg.Logging.OutputPath, "Log output path should not be empty")

			if cfg.DB.DatabaseURL != "" {
				assert.NotEmpty(t, cfg.DB.DatabaseURL, "Database URL should not be empty when set")
			} else {
				assert.True(t, cfg.DB.Host != "" || cfg.DB.ConnectionName != "", "Either host or connection name should be set")
				assert.NotEmpty(t, cfg.DB.User, "Database user should not be empty")
				assert.NotEmpty(t, cfg.DB.Password, "Database password should not be empty")
				assert.NotEmpty(t, cfg.DB.DBName, "Database name should not be empty")
			}
		})
	}
}

// TestEnvironmentOverrides tests environment variable overrides
func TestEnvironmentOverrides(t *testing.T) {
	t.Setenv("ENVIRONMENT", "staging")
	t.Setenv("DB_HOST", "test-host")
	t.Setenv("DB_USER", "test-user")
	t.Setenv("DB_PASSWORD", "test-password")
	t.Setenv("DB_NAME", "test-db")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("SERVER_PORT", "9090")

	cfg, err := config.LoadConfig("test-data/staging.yml")
	require.NoError(t, err, "Config should load successfully")

	assert.Equal(t, config.Staging, cfg.Environment, "Environment should be overridden")
	assert.Equal(t, "test-host", cfg.DB.Host, "DB host should be overridden")
	assert.Equal(t, "test-user", cfg.DB.User, "DB user should be overridden")
	assert.Equal(t, "test-password", cfg.DB.Password, "DB password should be overridden")
	assert.Equal(t, "test-db", cfg.DB.DBName, "DB name should be overridden")
	assert.Equal(t, "debug", cfg.Logging.Level, "Log level should be overridden")
	assert.Equal(t, 9090, cfg.Server.Port, "Server port should be overridden")
}

// TestDatabaseURLFormats tests different DATABASE_URL formats
func TestDatabaseURLFormats(t *testing.T) {
	formats := []string{
		"postgres://user:pass@host:5432/dbname",
		"postgresql://user:pass@host:5432/dbname",
		"postgres://user:pass@host:5432/dbname?sslmode=require",
		"postgresql://user:pass@host:5432/dbname?sslmode=disable",
		"postgres://user:pass@host:5432/dbname?sslmode=require&connect_timeout=10",
	}

	for i, format := range formats {
		t.Run(fmt.Sprintf("format_%d_%s", i+1, format), func(t *testing.T) {
			t.Setenv("ENVIRONMENT", "production")
			t.Setenv("DATABASE_URL", format)

			cfg, err := config.LoadConfig("test-data/production.yml")
			require.NoError(t, err, "Config should load successfully string format: %s", format)
			require.NotNil(t, cfg)
			assert.Equal(t, format, cfg.DB.DatabaseURL)
		})
	}
}

// TestCriticalFailureScenarios tests real-world failure scenarios that cause service startup failures
func TestCriticalFailureScenarios(t *testing.T) {
	tests := []struct {
		name            string
		configFile      string
		envVars         map[string]string
		expectErrSubstr []string
	}{
		{
			name:       "invalid_server_port_causes_tcp_probe_failure",
			configFile: "test-data/local.yml",
			envVars: map[string]string{
				"ENVIRONMENT": "local",
				"DB_PORT":     "99999",
			},
			expectErrSubstr: []string{"DB_PORT must be 1–65535"},
		},
		{
			name:       "missing_database_connection_causes_startup_failure",
			configFile: "test-data/staging.yml",
			envVars: map[string]string{
				"ENVIRONMENT": "staging",
			},
			expectErrSubstr: []string{
				"either DB_HOST or DB_CONNECTION_NAME is required",
				"DB_USER is required",
				"DB_PASSWORD is required",
				"DB_NAME is required",
			},
		},
		{
			name:       "invalid_database_port_causes_connection_failure",
			configFile: "test-data/local.yml",
			envVars: map[string]string{
				"ENVIRONMENT": "local",
				"DB_PORT":     "99999",
			},
			expectErrSubstr: []string{"DB_PORT must be 1–65535"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			cfg, err := config.LoadConfig(tt.configFile)

			if len(tt.expectErrSubstr) > 0 {
				require.Error(t, err)
				assert.Nil(t, cfg)
				errStr := err.Error()
				for _, substr := range tt.expectErrSubstr {
					assert.Contains(t, errStr, substr)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
			}
		})
	}
}
