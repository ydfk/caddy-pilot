package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Jwt      JwtConfig
	Database DatabaseConfig
}

type AppConfig struct {
	Port string `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type JwtConfig struct {
	Secret     string `mapstructure:"secret"`
	Expiration int    `mapstructure:"expiration"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	Path   string `mapstructure:"path"`
	DSN    string `mapstructure:"dsn"`
}

func (c DatabaseConfig) DriverName() string {
	driver := strings.TrimSpace(strings.ToLower(c.Driver))
	switch driver {
	case "", "sqlite", "sqlite3":
		return "sqlite"
	case "postgres", "postgresql":
		return "postgres"
	case "mysql":
		return "mysql"
	default:
		return driver
	}
}

var Current Config
var IsProduction bool

func Init() error {
	config, err := loadConfig("config")
	if err != nil {
		return err
	}

	Current = config
	IsProduction = Current.App.Env == "production"

	return nil
}

func loadConfig(configDir string) (Config, error) {
	configPath := filepath.Join(configDir, "config.yaml")
	configLoader := viper.New()

	if err := readConfig(configLoader, configPath); err != nil {
		return Config{}, err
	}

	config, err := unmarshalConfig(configLoader)
	if err != nil {
		return Config{}, err
	}

	env := strings.TrimSpace(strings.ToLower(config.App.Env))
	overridePaths := buildOverridePaths(configDir, env)
	for _, overridePath := range overridePaths {
		if err := mergeConfigIfExists(configLoader, overridePath); err != nil {
			return Config{}, err
		}
	}

	return unmarshalConfig(configLoader)
}

func buildOverridePaths(configDir string, env string) []string {
	overridePaths := make([]string, 0, 3)
	if env != "" {
		overridePaths = append(overridePaths, filepath.Join(configDir, fmt.Sprintf("config.%s.yaml", env)))
	}

	overridePaths = append(overridePaths, filepath.Join(configDir, "config.local.yaml"))
	if env != "" {
		overridePaths = append(overridePaths, filepath.Join(configDir, fmt.Sprintf("config.%s.local.yaml", env)))
	}

	return overridePaths
}

func readConfig(configLoader *viper.Viper, configPath string) error {
	configLoader.SetConfigFile(configPath)
	if err := configLoader.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败 %s: %w", configPath, err)
	}

	return nil
}

func mergeConfigIfExists(configLoader *viper.Viper, configPath string) error {
	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("检查配置文件失败 %s: %w", configPath, err)
	}

	configLoader.SetConfigFile(configPath)
	if err := configLoader.MergeInConfig(); err != nil {
		return fmt.Errorf("合并配置文件失败 %s: %w", configPath, err)
	}

	return nil
}

func unmarshalConfig(configLoader *viper.Viper) (Config, error) {
	var config Config
	if err := configLoader.Unmarshal(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}
