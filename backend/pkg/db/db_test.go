package db

import (
	"path/filepath"
	"testing"

	"go-fiber-starter/pkg/config"
)

func TestNewDialectorSQLite(t *testing.T) {
	t.Parallel()

	databaseConfig := config.DatabaseConfig{
		Driver: "sqlite",
		Path:   filepath.Join(t.TempDir(), "test.sqlite"),
	}

	dialector, err := newDialector(databaseConfig)
	if err != nil {
		t.Fatalf("newDialector returned error: %v", err)
	}
	if dialector.Name() != "sqlite" {
		t.Fatalf("dialector name %s != sqlite", dialector.Name())
	}
}

func TestNewDialectorPostgresAlias(t *testing.T) {
	t.Parallel()

	dialector, err := newDialector(config.DatabaseConfig{
		Driver: "postgresql",
		DSN:    "host=127.0.0.1 user=postgres password=postgres dbname=test port=5432 sslmode=disable",
	})
	if err != nil {
		t.Fatalf("newDialector returned error: %v", err)
	}
	if dialector.Name() != "postgres" {
		t.Fatalf("dialector name %s != postgres", dialector.Name())
	}
}

func TestNewDialectorMySQLEmptyDSN(t *testing.T) {
	t.Parallel()

	_, err := newDialector(config.DatabaseConfig{Driver: "mysql"})
	if err == nil {
		t.Fatal("newDialector expected error, got nil")
	}
}

func TestNewDialectorUnsupportedDriver(t *testing.T) {
	t.Parallel()

	_, err := newDialector(config.DatabaseConfig{Driver: "sqlserver"})
	if err == nil {
		t.Fatal("newDialector expected error, got nil")
	}
}
