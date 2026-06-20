package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewCaddyClientUsesDefaultAndTimeout(t *testing.T) {
	t.Setenv("CADDY_ADMIN_API", "")
	client := NewCaddyClient()
	if client.AdminAPI != DefaultCaddyAdminAPI {
		t.Fatalf("默认 Admin API 为 %s", client.AdminAPI)
	}
	if client.HTTPClient.Timeout != 5*time.Second {
		t.Fatalf("HTTP 超时为 %s", client.HTTPClient.Timeout)
	}
}

func TestCaddyClientGetAndLoadConfig(t *testing.T) {
	var loaded string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/config/":
			_, _ = writer.Write([]byte(`{"apps":{}}`))
		case "/load":
			if request.Method != http.MethodPost || request.Header.Get("Content-Type") != "application/json" {
				t.Errorf("发布请求不正确: %s %s", request.Method, request.Header.Get("Content-Type"))
			}
			body, _ := io.ReadAll(request.Body)
			loaded = string(body)
			writer.WriteHeader(http.StatusOK)
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	client := &CaddyClient{AdminAPI: server.URL, HTTPClient: server.Client()}
	config, err := client.GetConfig(context.Background())
	if err != nil || string(config) != `{"apps":{}}` {
		t.Fatalf("读取配置失败: %v, %s", err, config)
	}
	if err := client.LoadConfig(context.Background(), []byte(`{"admin":{}}`)); err != nil {
		t.Fatalf("发布配置失败: %v", err)
	}
	if loaded != `{"admin":{}}` {
		t.Fatalf("发布载荷不正确: %s", loaded)
	}
}

func TestCaddyClientReturnsClearStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, "配置无效", http.StatusBadRequest)
	}))
	defer server.Close()

	client := &CaddyClient{AdminAPI: server.URL, HTTPClient: server.Client()}
	err := client.GetStatus(context.Background())
	if err == nil || !strings.Contains(err.Error(), "400") || !strings.Contains(err.Error(), "配置无效") {
		t.Fatalf("错误信息不清晰: %v", err)
	}
}
