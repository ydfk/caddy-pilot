package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCaddyVersionServiceCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("User-Agent") != "CaddyPilot" {
			t.Fatal("缺少 User-Agent")
		}
		_, _ = writer.Write([]byte(`{"tag_name":"v2.11.0","html_url":"https://example.com/release"}`))
	}))
	defer server.Close()

	service := NewCaddyVersionService()
	service.ReleaseAPI = server.URL
	service.HTTPClient = server.Client()
	service.currentVersion = func(context.Context) (string, error) { return "2.10.0", nil }

	info, err := service.Check(context.Background())
	if err != nil {
		t.Fatalf("检查版本失败: %v", err)
	}
	if info.CurrentVersion != "2.10.0" || info.LatestVersion != "2.11.0" || !info.UpdateAvailable {
		t.Fatalf("版本信息不正确: %+v", info)
	}
	if !strings.Contains(info.UpdateCommand, "CADDY_VERSION='2.11.0'") {
		t.Fatalf("更新命令不正确: %s", info.UpdateCommand)
	}
}

func TestCaddyVersionServiceKeepsCurrentWhenReleaseUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	service := NewCaddyVersionService()
	service.ReleaseAPI = server.URL
	service.HTTPClient = server.Client()
	service.currentVersion = func(context.Context) (string, error) { return "2.10.0", nil }

	info, err := service.Check(context.Background())
	if err != nil || info.CurrentVersion != "2.10.0" || info.ErrorMessage == "" {
		t.Fatalf("离线版本信息不正确: %+v, %v", info, err)
	}
}

func TestCaddyVersionServiceSupportsCustomResponseAndUpdateURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"version":"v2.12.0","update_url":"https://mirror.example/caddy"}`))
	}))
	defer server.Close()
	t.Setenv("CADDY_UPDATE_URL", "https://download.example/caddy")

	service := NewCaddyVersionService()
	service.ReleaseAPI = server.URL
	service.HTTPClient = server.Client()
	service.currentVersion = func(context.Context) (string, error) { return "2.10.0", nil }

	info, err := service.Check(context.Background())
	if err != nil || info.LatestVersion != "2.12.0" || info.ReleaseURL != "https://download.example/caddy" {
		t.Fatalf("自定义版本服务结果不正确: %+v, %v", info, err)
	}
}

func TestCheckCaddyRuntimeRejectsMissingBinary(t *testing.T) {
	t.Setenv("CADDY_BINARY", "caddypilot-missing-caddy-binary")
	if _, err := CheckCaddyRuntime(context.Background()); err == nil {
		t.Fatal("缺少 Caddy 可执行文件时应返回错误")
	}
}
