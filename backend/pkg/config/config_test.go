package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigBaseOnly(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	writeConfigFile(t, configDir, "config.yaml", `
app:
  port: "25610"
  env: "development"
jwt:
  secret: "base"
  expiration: 604800
database:
  driver: "sqlite"
  path: "data/db.sqlite"
  dsn: ""
`)

	config, err := loadConfig(configDir)
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}

	if config.App.Port != "25610" {
		t.Fatalf("config.App.Port %s != 25610", config.App.Port)
	}
	if config.Jwt.Secret != "base" {
		t.Fatalf("config.Jwt.Secret %s != base", config.Jwt.Secret)
	}
}

func TestLoadConfigOverrideOrder(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	writeConfigFile(t, configDir, "config.yaml", `
app:
  port: "25610"
  env: "development"
jwt:
  secret: "base"
  expiration: 604800
database:
  driver: "sqlite"
  path: "data/db.sqlite"
  dsn: ""
`)
	writeConfigFile(t, configDir, "config.development.yaml", `
app:
  port: "3000"
`)
	writeConfigFile(t, configDir, "config.local.yaml", `
jwt:
  secret: "local"
`)
	writeConfigFile(t, configDir, "config.development.local.yaml", `
app:
  port: "4000"
`)

	config, err := loadConfig(configDir)
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}

	if config.App.Port != "4000" {
		t.Fatalf("config.App.Port %s != 4000", config.App.Port)
	}
	if config.Jwt.Secret != "local" {
		t.Fatalf("config.Jwt.Secret %s != local", config.Jwt.Secret)
	}
}

func TestLoadConfigMissingBaseConfig(t *testing.T) {
	t.Parallel()

	_, err := loadConfig(t.TempDir())
	if err == nil {
		t.Fatal("loadConfig expected error, got nil")
	}
}

func writeConfigFile(t *testing.T, configDir string, fileName string, content string) {
	t.Helper()

	configPath := filepath.Join(configDir, fileName)
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
}
