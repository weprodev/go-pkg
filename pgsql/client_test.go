package pgsql_test

import (
	"testing"

	"github.com/weprodev/go-pkg/pgsql"
)

// In a unit test environment without a real Postgres instance,
// NewPgClient will fail at Ping(). We can test just the config
// and helper functionality.

func TestDefaultPgConfig(t *testing.T) {
	cfg := pgsql.DefaultPgConfig()
	if cfg.Host != "localhost" {
		t.Errorf("expected localhost, got %s", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("expected 5432, got %d", cfg.Port)
	}
	if cfg.MaxOpenConns != 25 {
		t.Errorf("expected 25, got %d", cfg.MaxOpenConns)
	}
}

func TestNewPgClient_FailsToConnect(t *testing.T) {
	// Without a real database on port 12345, this should fail at Ping
	cfg := pgsql.DefaultPgConfig()
	cfg.Port = 12345

	_, err := pgsql.NewPgClient(cfg)
	if err == nil {
		t.Error("expected connection to fail but it succeeded")
	}
}

// Since we can't easily mock *sql.DB for RunInTransaction in a black-box
// test without an interface or sqlmock, we will limit the pgsql testing
// to what we can do without a DB for now.
// A real test would spin up testcontainers-go or require a local DB.

// To improve coverage slightly, we can test that GetDB returns the provided tx
// if one is injected into the context, but GetDB requires an actual PgClient
// instance which requires a successful DB connection to initialize.
