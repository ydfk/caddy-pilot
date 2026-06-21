package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"
	"go-fiber-starter/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		logger.Error("CaddyPilot 已退出: %v", err)
		os.Exit(1)
	}
}

func run() error {
	if err := logger.Init(); err != nil {
		return err
	}

	if err := config.Init(); err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if err := db.Init(); err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	manager := service.NewCaddyManager()
	runtimeInfo, err := manager.Start(ctx)
	if err != nil {
		return fmt.Errorf("托管 Caddy 启动失败: %w", err)
	}
	service.SetManagedCaddy(manager)
	logger.Info("托管 Caddy 已启动: %s (%s)", runtimeInfo.Version, runtimeInfo.BinaryPath)

	app := newApp()
	serverDone := make(chan error, 1)
	address := config.Current.App.ListenAddress()
	go func() {
		serverDone <- app.Listen(address)
	}()
	logger.Info("后端服务已启动: http://%s", address)

	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-manager.Done():
	case runErr = <-serverDone:
		if runErr != nil {
			runErr = fmt.Errorf("后端服务退出: %w", runErr)
		}
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownContext); err != nil && runErr == nil {
		runErr = fmt.Errorf("关闭后端服务失败: %w", err)
	}
	if err := manager.Stop(shutdownContext); err != nil && runErr == nil {
		runErr = fmt.Errorf("关闭 Caddy 失败: %w", err)
	}
	return runErr
}
