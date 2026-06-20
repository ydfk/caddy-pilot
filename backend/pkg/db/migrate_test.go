package db

import (
	"testing"

	"go-fiber-starter/internal/model/caddynode"
	"go-fiber-starter/internal/model/proxysite"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAutoMigrateCreatesBusinessTablesAndLocalNode(t *testing.T) {
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

	for _, table := range []string{"users", "proxy_sites", "config_versions", "caddy_nodes"} {
		if !database.Migrator().HasTable(table) {
			t.Fatalf("缺少数据表 %s", table)
		}
	}
	for _, column := range []string{"enable_https", "force_https"} {
		if !database.Migrator().HasColumn(&proxysite.ProxySite{}, column) {
			t.Fatalf("proxy_sites 缺少字段 %s", column)
		}
	}

	var node caddynode.CaddyNode
	if err := database.Where("name = ?", "local").First(&node).Error; err != nil {
		t.Fatalf("读取默认 Caddy 节点失败: %v", err)
	}
	if node.AdminAPI != "http://127.0.0.1:2019" || !node.Enabled {
		t.Fatalf("默认 Caddy 节点配置不正确: %+v", node)
	}
}
