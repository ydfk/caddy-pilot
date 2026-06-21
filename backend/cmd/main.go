package main

import (
	"context"

	"go-fiber-starter/internal/service"
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

	runtime, err := service.CheckCaddyRuntime(context.Background())
	if err != nil {
		if config.IsProduction {
			logger.Fatal("Caddy 运行时检查失败: %v", err)
		}
		logger.Warn("Caddy 运行时检查失败，本地开发继续启动: %v", err)
	} else {
		logger.Info("Caddy 运行时检查通过: %s (%s)", runtime.Version, runtime.BinaryPath)
	}

	if err := db.Init(); err != nil {
		logger.Fatal("初始化数据库失败: %v", err)
	}

	api()
}
