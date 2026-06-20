package db

import (
	"errors"
	"fmt"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/logger"
	"go-fiber-starter/pkg/util"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

var DB *gorm.DB

func Init() error {
	db, err := openDatabase(config.Current.Database)
	if err != nil {
		return err
	}

	DB = db
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	return nil
}

func openDatabase(databaseConfig config.DatabaseConfig) (*gorm.DB, error) {
	dialector, err := newDialector(databaseConfig)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:         zapgorm2.New(logger.Logger.Desugar()),
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	return db, nil
}

func newDialector(databaseConfig config.DatabaseConfig) (gorm.Dialector, error) {
	switch databaseConfig.DriverName() {
	case "sqlite":
		return newSQLiteDialector(databaseConfig)
	case "postgres":
		return newPostgresDialector(databaseConfig)
	case "mysql":
		return newMySQLDialector(databaseConfig)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", databaseConfig.Driver)
	}
}

func newSQLiteDialector(databaseConfig config.DatabaseConfig) (gorm.Dialector, error) {
	path := strings.TrimSpace(databaseConfig.Path)
	if path == "" {
		return nil, errors.New("sqlite 数据库 path 不能为空")
	}
	if err := util.EnsureDir(path); err != nil {
		logger.Error("创建 SQLite 数据库目录失败: %v", err)
		return nil, err
	}

	return sqlite.Open(path), nil
}

func newPostgresDialector(databaseConfig config.DatabaseConfig) (gorm.Dialector, error) {
	dsn := strings.TrimSpace(databaseConfig.DSN)
	if dsn == "" {
		return nil, errors.New("postgres 数据库 dsn 不能为空")
	}

	return postgres.Open(dsn), nil
}

func newMySQLDialector(databaseConfig config.DatabaseConfig) (gorm.Dialector, error) {
	dsn := strings.TrimSpace(databaseConfig.DSN)
	if dsn == "" {
		return nil, errors.New("mysql 数据库 dsn 不能为空")
	}

	return mysql.Open(dsn), nil
}
