package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const CaddyLatestReleaseAPI = "https://api.github.com/repos/caddyserver/caddy/releases/latest"

type CaddyVersionInfo struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	BinaryPath      string
	VersionCheckURL string
	DownloadURL     string
	UpdateURL       string
	ReleaseURL      string
	ErrorMessage    string
}

type CaddyVersionService struct {
	Binary         string
	ReleaseAPI     string
	HTTPClient     *http.Client
	currentVersion func(context.Context) (string, error)
}

func NewCaddyVersionService() *CaddyVersionService {
	service := &CaddyVersionService{
		Binary:     environmentValue("CADDY_BINARY", "caddy"),
		ReleaseAPI: environmentValue("CADDY_VERSION_CHECK_URL", CaddyLatestReleaseAPI),
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
	service.currentVersion = service.readCurrentVersion
	return service
}

func (service *CaddyVersionService) Check(ctx context.Context) (CaddyVersionInfo, error) {
	current, err := service.currentVersion(ctx)
	if err != nil {
		return CaddyVersionInfo{}, err
	}

	info := CaddyVersionInfo{
		CurrentVersion:  current,
		VersionCheckURL: service.ReleaseAPI,
		DownloadURL:     environmentValue("CADDY_DOWNLOAD_URL", DefaultCaddyDownloadURL),
		UpdateURL:       strings.TrimSpace(os.Getenv("CADDY_UPDATE_URL")),
	}
	if path, lookupErr := exec.LookPath(service.Binary); lookupErr == nil {
		info.BinaryPath = path
	}
	latest, releaseURL, err := service.readLatestRelease(ctx)
	if err != nil {
		info.ErrorMessage = err.Error()
		return info, nil
	}

	info.LatestVersion = latest
	info.ReleaseURL = releaseURL
	if info.UpdateURL != "" {
		info.ReleaseURL = info.UpdateURL
	}
	info.UpdateAvailable = normalizeVersion(current) != normalizeVersion(latest)
	return info, nil
}

func (service *CaddyVersionService) readCurrentVersion(ctx context.Context) (string, error) {
	output, err := exec.CommandContext(ctx, service.Binary, "version").Output()
	if err != nil {
		return "", fmt.Errorf("读取 Caddy 版本失败: %w", err)
	}
	fields := strings.Fields(string(output))
	if len(fields) == 0 {
		return "", fmt.Errorf("Caddy 版本输出为空")
	}
	return normalizeVersion(fields[0]), nil
}

func (service *CaddyVersionService) readLatestRelease(ctx context.Context) (string, string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, service.ReleaseAPI, nil)
	if err != nil {
		return "", "", fmt.Errorf("创建 Caddy 版本请求失败: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "CaddyPilot")

	response, err := service.HTTPClient.Do(request)
	if err != nil {
		return "", "", fmt.Errorf("检查 Caddy 更新失败: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("检查 Caddy 更新失败: 版本服务返回 %d", response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", "", fmt.Errorf("读取 Caddy 版本响应失败: %w", err)
	}
	var release struct {
		TagName   string `json:"tag_name"`
		HTMLURL   string `json:"html_url"`
		Version   string `json:"version"`
		UpdateURL string `json:"update_url"`
	}
	if err := json.Unmarshal(body, &release); err != nil {
		return "", "", fmt.Errorf("解析 Caddy 版本响应失败: %w", err)
	}
	version := release.TagName
	if strings.TrimSpace(version) == "" {
		version = release.Version
	}
	updateURL := release.HTMLURL
	if strings.TrimSpace(updateURL) == "" {
		updateURL = release.UpdateURL
	}
	if strings.TrimSpace(version) == "" {
		return "", "", fmt.Errorf("Caddy 最新版本为空")
	}
	return normalizeVersion(version), updateURL, nil
}

func normalizeVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}

func environmentValue(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
