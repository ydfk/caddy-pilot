package service

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCaddyVersionRequestErrorExplainsContainerDNS(t *testing.T) {
	err := caddyVersionRequestError("api.github.com", &net.DNSError{Name: "api.github.com", Err: "server misbehaving"})
	if !errors.As(err, new(*net.DNSError)) || !strings.Contains(err.Error(), "CADDYPILOT_DNS_SERVER") {
		t.Fatalf("DNS 错误提示不完整: %v", err)
	}
}

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
	if !strings.Contains(info.DownloadURL, "{version}") {
		t.Fatalf("下载地址模板不正确: %s", info.DownloadURL)
	}
}

func TestCaddyVersionServiceAcceptsGitHubAssetsArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{
			"tag_name":"v2.12.0",
			"html_url":"https://github.com/caddyserver/caddy/releases/tag/v2.12.0",
			"assets":[{"name":"caddy_2.12.0_linux_amd64.tar.gz","browser_download_url":"https://example.com/caddy.tar.gz"}]
		}`))
	}))
	defer server.Close()
	service := NewCaddyVersionService()
	service.ReleaseAPI = server.URL
	service.HTTPClient = server.Client()
	service.currentVersion = func(context.Context) (string, error) { return "2.11.0", nil }

	info, err := service.Check(context.Background())
	if err != nil || info.ErrorMessage != "" || info.LatestVersion != "2.12.0" {
		t.Fatalf("GitHub Release 响应解析失败: %+v, %v", info, err)
	}
	if info.DownloadURL != DefaultCaddyDownloadURL {
		t.Fatalf("GitHub Release 资产不应覆盖定制构建下载源: %s", info.DownloadURL)
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
	service := NewCaddyVersionService()
	service.ReleaseAPI = server.URL
	service.HTTPClient = server.Client()
	service.currentVersion = func(context.Context) (string, error) { return "2.10.0", nil }

	info, err := service.Check(context.Background())
	if err != nil || info.LatestVersion != "2.12.0" || info.ReleaseURL != "https://mirror.example/caddy" {
		t.Fatalf("自定义版本服务结果不正确: %+v, %v", info, err)
	}
}

func TestCaddyVersionServiceReadsRuntimeManifest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{
			"version":"2.11.4",
			"release_tag":"v1.2.3",
			"assets":{"windows_amd64":{"url":"https://mirror.example/caddy.zip"}},
			"sha512_url":"https://mirror.example/sha512sums.txt"
		}`))
	}))
	defer server.Close()
	service := NewCaddyVersionService()
	service.ReleaseAPI = server.URL
	service.HTTPClient = server.Client()
	service.GOOS = "windows"
	service.GOARCH = "amd64"
	service.currentVersion = func(context.Context) (string, error) { return "2.10.0", nil }

	info, err := service.Check(context.Background())
	if err != nil || info.LatestVersion != "2.11.4" || info.DownloadURL != "https://mirror.example/caddy.zip" || info.ChecksumURL != "https://mirror.example/sha512sums.txt" {
		t.Fatalf("运行时 manifest 解析失败: %+v, %v", info, err)
	}
}
