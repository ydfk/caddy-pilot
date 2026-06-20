package main

import (
	"go-fiber-starter/internal/api/auth"
	"go-fiber-starter/internal/api/caddy"
	"go-fiber-starter/internal/api/configversion"
	"go-fiber-starter/internal/api/dashboard"
	"go-fiber-starter/internal/api/proxysite"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/logger"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func api() {
	app := newApp()
	if err := app.Listen(":" + config.Current.App.Port); err != nil {
		logger.Fatal("启动服务器失败: %v", err)
	} else {
		logger.Info("服务器启动成功: http://127.0.0.1:%v", config.Current.App.Port)
		logger.Info("API 文档: http://127.0.0.1:%v/docs", config.Current.App.Port)
	}
}

func newApp() *fiber.App {
	app := fiber.New()
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(fiberLogger.New(fiberLogger.Config{
		Format: "${ip} ${status} ${latency} ${method} ${path}\n",
		Stream: logger.GetFiberLogWriter(),
	}))

	humaConfig := huma.DefaultConfig("Go Fiber Starter API", "1.0.0")
	humaConfig.Info.Description = "基于 Fiber v3 与 Huma 的 OpenAPI 3.1 code-first API"
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		auth.BearerAuthScheme: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}
	humaAPI := humafiber.New(app, humaConfig)
	humaAPI.UseMiddleware(auth.NewAuthMiddleware(humaAPI))
	auth.RegisterRoutes(humaAPI)
	proxysite.RegisterRoutes(humaAPI)
	caddy.RegisterRoutes(humaAPI)
	configversion.RegisterRoutes(humaAPI)
	dashboard.RegisterRoutes(humaAPI)
	return app
}
