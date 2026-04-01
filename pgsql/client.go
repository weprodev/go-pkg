package pgsql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PgConfig holds all parameters needed to open a PostgreSQL connection.
type PgConfig struct {
	// Host is the hostname or IP address for TCP connections.
	Host string
	// ConnectionName is an instance identifier used with Unix-socket connections
	// (e.g. a managed database proxy socket path). When set, the socket path
	// is derived as /path/to/socket/<ConnectionName>.
	ConnectionName  string
	DatabaseURL     string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DefaultPgConfig returns a PgConfig with sane defaults for local development.
func DefaultPgConfig() PgConfig {
	return PgConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "",
		DBName:          "postgres",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

// PgClient wraps *sql.DB and adds transaction-context helpers.
type PgClient struct {
	*sql.DB
	config PgConfig
}

// NewPgClient opens a PostgreSQL connection using the provided config.
// Defaults are applied for any zero-value pool settings.
func NewPgClient(config PgConfig) (*PgClient, error) {
	def := DefaultPgConfig()
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = def.MaxOpenConns
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = def.MaxIdleConns
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = def.ConnMaxLifetime
	}
	if config.SSLMode == "" {
		config.SSLMode = def.SSLMode
	}

	connStr := buildConnString(config)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &PgClient{DB: db, config: config}, nil
}

// buildConnString selects the appropriate connection string format.
func buildConnString(config PgConfig) string {
	switch {
	case config.DatabaseURL != "":
		return config.DatabaseURL

	case config.ConnectionName != "":
		// Managed database Unix-socket path (e.g. /cloudsql/<instance>).
		socketPath := fmt.Sprintf("/cloudsql/%s", config.ConnectionName)
		return fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s sslmode=%s",
			socketPath, config.User, config.Password, config.DBName, config.SSLMode,
		)

	case isUnixSocket(config.Host):
		return fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.User, config.Password, config.DBName, config.SSLMode,
		)

	default:
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
		)
	}
}

// isUnixSocket reports whether host looks like a Unix socket path.
func isUnixSocket(host string) bool {
	return len(host) > 0 && host[0] == '/'
}

// Close closes the underlying database connection.
func (c *PgClient) Close() error { return c.DB.Close() }

// GetConfig returns the config used to open this client.
func (c *PgClient) GetConfig() PgConfig { return c.config }

// txKey is the context key for an in-progress transaction.
type txKey struct{}

// DBTX is the minimal interface satisfied by both *sql.DB and *sql.Tx,
// allowing repository code to be transaction-agnostic.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// GetDB returns the active *sql.Tx from ctx if one exists, otherwise the
// underlying *sql.DB. This lets repositories call GetDB transparently and
// participate in outer transactions without extra plumbing.
func (c *PgClient) GetDB(ctx context.Context) DBTX {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return c.DB
}

// RunInTransaction executes fn inside a database transaction.
// If ctx already carries a transaction, fn is called directly (no nesting).
// The transaction is committed on success and rolled back on error or panic.
func (c *PgClient) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if _, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return fn(ctx)
	}

	tx, err := c.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("commit transaction: %w", commitErr)
			}
		}
	}()

	err = fn(context.WithValue(ctx, txKey{}, tx))
	return err
}
