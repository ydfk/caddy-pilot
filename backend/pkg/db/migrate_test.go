package db

import (
	"testing"

	"go-fiber-starter/internal/model/basicauth"
	"go-fiber-starter/internal/model/proxysite"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAutoMigrateCreatesBusinessTables(t *testing.T) {
	previousDB := DB
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	DB = database
	t.Cleanup(func() { DB = previousDB })

	if err := autoMigrate(); err != nil {
		t.Fatalf("执行数据库迁移失败: %v", err)
	}

	for _, table := range []string{"users", "proxy_sites", "config_versions", "credentials", "settings"} {
		if !database.Migrator().HasTable(table) {
			t.Fatalf("缺少数据表 %s", table)
		}
	}
	for _, column := range []string{"enable_https", "force_https"} {
		if !database.Migrator().HasColumn(&proxysite.ProxySite{}, column) {
			t.Fatalf("proxy_sites 缺少字段 %s", column)
		}
	}

	if !database.Migrator().HasColumn(&basicauth.Credential{}, "password_hash") {
		t.Fatal("credentials 缺少 password_hash 字段")
	}
}
