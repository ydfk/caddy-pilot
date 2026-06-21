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
