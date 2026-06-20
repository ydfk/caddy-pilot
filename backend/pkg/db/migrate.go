package db

import model "go-fiber-starter/internal/model/user"

// autoMigrate 自动迁移数据库表
func autoMigrate() error {
	return DB.AutoMigrate(
		&model.User{},
	)
}
