package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAPIContainsCaddyPilotRoutes(t *testing.T) {
	app := newApp()
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil))
	if err != nil {
		t.Fatalf("读取 OpenAPI 失败: %v", err)
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("读取 OpenAPI 响应失败: %v", err)
	}
	for _, route := range []string{
		"/api/proxy-sites:", "/api/caddy/publish:",
		"/api/caddy/settings:",
		"/api/caddy/upload:", "/api/caddy/update-tasks/current:",
		"/api/logs:",
		"/api/system/info:",
		"/api/config-versions/{id}/rollback:", "/api/dashboard/summary:",
	} {
		if !strings.Contains(string(payload), route) {
			t.Fatalf("OpenAPI 缺少路由 %s", route)
		}
	}
}
