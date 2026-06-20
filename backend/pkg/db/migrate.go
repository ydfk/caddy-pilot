package db

import (
	"go-fiber-starter/internal/model/caddynode"
	"go-fiber-starter/internal/model/configversion"
	"go-fiber-starter/internal/model/proxysite"
	"go-fiber-starter/internal/model/user"

	"gorm.io/gorm"
)

// autoMigrate 自动迁移数据库表
func autoMigrate() error {
	if err := DB.AutoMigrate(
		&user.User{},
		&proxysite.ProxySite{},
		&configversion.ConfigVersion{},
		&caddynode.CaddyNode{},
	); err != nil {
		return err
	}

	return seedLocalCaddyNode(DB)
}

func seedLocalCaddyNode(database *gorm.DB) error {
	node := caddynode.CaddyNode{Name: "local"}
	return database.Where("name = ?", node.Name).FirstOrCreate(&node, caddynode.CaddyNode{
		Name:     "local",
		AdminAPI: "http://127.0.0.1:2019",
		Enabled:  true,
	}).Error

}
