package service

import (
	"bufio"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	caddyDownloadAttempts = 3
	maxErrorResponseBytes = 2048
)

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type CaddyUpdateProgress struct {
	Stage           string
	Attempt         int
	EffectiveURL    string
	DownloadedBytes int64
	TotalBytes      int64
	HTTPStatus      int
}

type CaddyDownloadError struct {
	Stage           string
	Attempt         int
	EffectiveURL    string
	HTTPStatus      int
	DownloadedBytes int64
	ResponseBody    string
	Cause           error
}

func (downloadErr *CaddyDownloadError) Error() string {
	parts := []string{fmt.Sprintf("Caddy 下载在%s阶段失败", downloadErr.Stage)}
	if downloadErr.Attempt > 0 {
		parts = append(parts, fmt.Sprintf("第 %d/%d 次尝试", downloadErr.Attempt, caddyDownloadAttempts))
	}
	if downloadErr.EffectiveURL != "" {
		parts = append(parts, "地址 "+downloadErr.EffectiveURL)
	}
	if downloadErr.HTTPStatus > 0 {
		parts = append(parts, fmt.Sprintf("HTTP %d", downloadErr.HTTPStatus))
	}
	parts = append(parts, fmt.Sprintf("已下载 %d 字节", downloadErr.DownloadedBytes))
	if downloadErr.ResponseBody != "" {
		parts = append(parts, "响应 "+downloadErr.ResponseBody)
	}
	if downloadErr.Cause != nil {
		parts = append(parts, downloadErr.Cause.Error())
	}
	return strings.Join(parts, "；")
}

func (downloadErr *CaddyDownloadError) Unwrap() error {
	return downloadErr.Cause
}

func newCaddyHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Minute}
}

func (installer *CaddyInstaller) download(ctx context.Context, version, destination string) error {
	partial := destination + ".part"
	var lastErr error
	for attempt := 1; attempt <= caddyDownloadAttempts; attempt++ {
		lastErr = installer.downloadAttempt(ctx, version, partial, attempt)
		if lastErr == nil {
			return replaceFile(partial, destination)
		}
		if attempt == caddyDownloadAttempts || ctx.Err() != nil {
			break
		}
		if err := waitForRetry(ctx, installer.retryDelay(attempt)); err != nil {
			return err
		}
	}
	return lastErr
}

func (installer *CaddyInstaller) downloadAttempt(ctx context.Context, version, partial string, attempt int) error {
	downloaded := fileSize(partial)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, installer.downloadURL(version), nil)
	if err != nil {
		return installer.downloadError("request", attempt, downloaded, 0, "", "", err)
	}
	request.Header.Set("User-Agent", "CaddyPilot")
	if downloaded > 0 {
		request.Header.Set("Range", fmt.Sprintf("bytes=%d-", downloaded))
	}
	installer.report(CaddyUpdateProgress{Stage: "downloading", Attempt: attempt, EffectiveURL: request.URL.String(), DownloadedBytes: downloaded})
	response, err := installer.HTTPClient.Do(request)
	if err != nil {
		return installer.downloadError("request", attempt, downloaded, 0, request.URL.String(), "", err)
	}
	defer response.Body.Close()
	effectiveURL := request.URL.String()
	if response.Request != nil && response.Request.URL != nil {
		effectiveURL = response.Request.URL.String()
	}
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusPartialContent {
		body := readErrorResponse(response.Body)
		return installer.downloadError("response", attempt, downloaded, response.StatusCode, effectiveURL, body, nil)
	}
	if response.StatusCode == http.StatusOK && downloaded > 0 {
		downloaded = 0
	}
	file, err := openPartialFile(partial, downloaded)
	if err != nil {
		return installer.downloadError("write", attempt, downloaded, response.StatusCode, effectiveURL, "", err)
	}
	total := response.ContentLength
	if total >= 0 {
		total += downloaded
	}
	reader := &downloadProgressReader{
		reader: response.Body, downloaded: downloaded, total: total,
		report: func(current, expected int64) {
			installer.report(CaddyUpdateProgress{
				Stage: "downloading", Attempt: attempt, EffectiveURL: effectiveURL,
				DownloadedBytes: current, TotalBytes: expected, HTTPStatus: response.StatusCode,
			})
		},
	}
	written, copyErr := io.Copy(file, io.LimitReader(reader, MaxCaddyUploadSize-downloaded+1))
	closeErr := file.Close()
	current := downloaded + written
	if current > MaxCaddyUploadSize {
		return installer.downloadError("write", attempt, current, response.StatusCode, effectiveURL, "", fmt.Errorf("Caddy 下载文件超过 128 MiB 限制"))
	}
	if copyErr != nil {
		return installer.downloadError("write", attempt, current, response.StatusCode, effectiveURL, "", copyErr)
	}
	if closeErr != nil {
		return installer.downloadError("write", attempt, current, response.StatusCode, effectiveURL, "", closeErr)
	}
	if total >= 0 && current != total {
		return installer.downloadError("write", attempt, current, response.StatusCode, effectiveURL, "", io.ErrUnexpectedEOF)
	}
	return nil
}

func (installer *CaddyInstaller) verifyChecksum(ctx context.Context, version, archivePath string) error {
	if strings.TrimSpace(installer.ChecksumURL) == "" {
		return fmt.Errorf("Caddy 安装包缺少必需的 SHA-512 清单地址")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, installer.checksumURL(version), nil)
	if err != nil {
		return fmt.Errorf("创建 Caddy 校验和请求失败: %w", err)
	}
	request.Header.Set("User-Agent", "CaddyPilot")
	response, err := installer.HTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("下载 Caddy 校验和失败: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("下载 Caddy 校验和失败: 校验服务返回 %d: %s", response.StatusCode, readErrorResponse(response.Body))
	}
	wantedName := installer.checksumAssetName(version)
	wantedHash := ""
	scanner := bufio.NewScanner(io.LimitReader(response.Body, 1<<20))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && strings.TrimPrefix(fields[len(fields)-1], "*") == wantedName {
			wantedHash = fields[0]
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取 Caddy 校验和失败: %w", err)
	}
	if wantedHash == "" {
		return fmt.Errorf("Caddy 校验和文件中缺少 %s", wantedName)
	}
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha512.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	if !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), wantedHash) {
		return fmt.Errorf("Caddy 下载文件 SHA-512 校验失败")
	}
	return nil
}

func (installer *CaddyInstaller) checksumAssetName(version string) string {
	parsed, err := url.Parse(installer.downloadURL(version))
	if err == nil {
		if name := filepath.Base(parsed.Path); name != "." && name != "/" && name != "" {
			return name
		}
	}
	return installer.archiveName(version)
}

func (installer *CaddyInstaller) downloadError(stage string, attempt int, downloaded int64, status int, effectiveURL, body string, cause error) error {
	installer.report(CaddyUpdateProgress{
		Stage: stage, Attempt: attempt, EffectiveURL: effectiveURL,
		DownloadedBytes: downloaded, HTTPStatus: status,
	})
	return &CaddyDownloadError{
		Stage: stage, Attempt: attempt, EffectiveURL: effectiveURL, HTTPStatus: status,
		DownloadedBytes: downloaded, ResponseBody: body, Cause: cause,
	}
}

func (installer *CaddyInstaller) retryDelay(attempt int) time.Duration {
	base := installer.RetryDelay
	if base <= 0 {
		base = time.Second
	}
	return base * time.Duration(1<<(attempt-1))
}

func openPartialFile(path string, offset int64) (*os.File, error) {
	flags := os.O_CREATE | os.O_WRONLY
	if offset == 0 {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_APPEND
	}
	return os.OpenFile(path, flags, 0o600)
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func readErrorResponse(reader io.Reader) string {
	payload, _ := io.ReadAll(io.LimitReader(reader, maxErrorResponseBytes+1))
	text := strings.TrimSpace(string(payload))
	if len(payload) > maxErrorResponseBytes {
		text = strings.TrimSpace(string(payload[:maxErrorResponseBytes])) + "…"
	}
	return text
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type downloadProgressReader struct {
	reader     io.Reader
	downloaded int64
	total      int64
	report     func(int64, int64)
}

func (reader *downloadProgressReader) Read(payload []byte) (int, error) {
	count, err := reader.reader.Read(payload)
	reader.downloaded += int64(count)
	reader.report(reader.downloaded, reader.total)
	return count, err
}
