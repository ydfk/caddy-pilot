package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const DefaultCaddyAdminAPI = "http://127.0.0.1:2019"

type CaddyAdmin interface {
	GetConfig(context.Context) ([]byte, error)
	LoadConfig(context.Context, []byte) error
	GetStatus(context.Context) error
}

type CaddyClient struct {
	AdminAPI   string
	HTTPClient *http.Client
}

func NewCaddyClient() *CaddyClient {
	adminAPI := strings.TrimSpace(os.Getenv("CADDY_ADMIN_API"))
	if adminAPI == "" {
		adminAPI = DefaultCaddyAdminAPI
	}
	return &CaddyClient{
		AdminAPI:   strings.TrimRight(adminAPI, "/"),
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (client *CaddyClient) GetConfig(ctx context.Context) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, client.AdminAPI+"/config/", nil)
	if err != nil {
		return nil, fmt.Errorf("创建 Caddy 配置请求失败: %w", err)
	}
	return client.do(request)
}

func (client *CaddyClient) LoadConfig(ctx context.Context, config []byte) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.AdminAPI+"/load", bytes.NewReader(config))
	if err != nil {
		return fmt.Errorf("创建 Caddy 发布请求失败: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if _, err := client.do(request); err != nil {
		return fmt.Errorf("发布 Caddy 配置失败: %w", err)
	}
	return nil
}

func (client *CaddyClient) GetStatus(ctx context.Context) error {
	if _, err := client.GetConfig(ctx); err != nil {
		return fmt.Errorf("Caddy Admin API 不可用: %w", err)
	}
	return nil
}

func (client *CaddyClient) do(request *http.Request) ([]byte, error) {
	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("请求 Caddy Admin API 失败: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("读取 Caddy Admin API 响应失败: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = response.Status
		}
		return nil, fmt.Errorf("Caddy Admin API 返回 %d: %s", response.StatusCode, message)
	}
	return body, nil
}
