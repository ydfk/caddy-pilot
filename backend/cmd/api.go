package main

import (
	"go-fiber-starter/internal/api/auth"
	"go-fiber-starter/internal/api/caddy"
	"go-fiber-starter/internal/api/configversion"
	"go-fiber-starter/internal/api/dashboard"
	"go-fiber-starter/internal/api/proxysite"
	"go-fiber-starter/pkg/logger"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

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
