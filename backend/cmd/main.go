/*
 * @Description: Copyright (c) ydfk. All rights reserved
 * @Author: ydfk
 * @Date: 2025-06-09 16:38:19
 * @LastEditors: ydfk
 * @LastEditTime: 2025-06-10 16:51:49
 */
package main

import (
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"
	"go-fiber-starter/pkg/logger"
)

func main() {
	if err := logger.Init(); err != nil {
		panic(err)
	}

	if err := config.Init(); err != nil {
		logger.Fatal("加载配置失败: %v", err)
	}

	if err := db.Init(); err != nil {
		logger.Fatal("初始化数据库失败: %v", err)
	}

	api()
}
