package pg

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Postgres struct {
	*sql.DB
}

var postgresInstance *Postgres

// GetClient returns the Postgres singleton instance
func GetClient() *Postgres {
	return postgresInstance
}

// SetClient sets the Postgres singleton instance
func SetClient(client *sql.DB) {
	postgresInstance = &Postgres{client}
}

// GetPostgresConfigFromEnv retrieves Postgres connection parameters from environment variables.
func getPostgresConfigFromEnv() (host, port, user, password, dbname, sslmode string) {
	host = getenvOrDefault("DB_HOST", "localhost")
	port = getenvOrDefault("DB_PORT", "5432")
	user = getenvOrDefault("DB_USER", "postgres")
	password = getenvOrDefault("DB_PASSWORD", "postgres")
	dbname = getenvOrDefault("DB_NAME", "postgres")
	sslmode = getenvOrDefault("DB_SSLMODE", "disable") // default for local/dev
	return
}

// getenvOrDefault returns the value of the environment variable if set, otherwise returns the default.
func getenvOrDefault(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

// ConnectAndInit connects to Postgres and initializes IMS tables if not present.
func PgConnect(ctx context.Context) (*sql.DB, error) {
	host, port, user, password, dbname, sslmode := getPostgresConfigFromEnv()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB connection: %w", err)
	}

	// Optional: Set connection pool configs
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	// Initialize IMS tables
	if err := initIMSTables(db); err != nil {
		return nil, fmt.Errorf("failed to initialize IMS tables: %w", err)
	}

	return db, nil
}

// initIMSTables creates tables if they do not exist
func initIMSTables(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS hubs (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			address TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS skus (
			id SERIAL PRIMARY KEY,
			tenant_id INT NOT NULL,
			seller_id INT NOT NULL,
			sku_code VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (tenant_id, seller_id, sku_code)
		);`,
		`CREATE TABLE IF NOT EXISTS inventory (
			id SERIAL PRIMARY KEY,
			hub_id INT NOT NULL REFERENCES hubs(id) ON DELETE CASCADE,
			sku_id INT NOT NULL REFERENCES skus(id) ON DELETE CASCADE,
			quantity INT NOT NULL DEFAULT 0,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (hub_id, sku_id)
		);`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed executing statement: %v, error: %w", stmt, err)
		}
	}

	return nil
}
