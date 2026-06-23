package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCaddyDownloadResumesPartialFile(t *testing.T) {
	payload := []byte("complete-caddy-archive")
	var rangeHeader string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		rangeHeader = request.Header.Get("Range")
		writer.Header().Set("Content-Length", fmt.Sprint(len(payload)-8))
		writer.WriteHeader(http.StatusPartialContent)
		_, _ = writer.Write(payload[8:])
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "caddy.zip")
	if err := os.WriteFile(destination+".part", payload[:8], 0o600); err != nil {
		t.Fatal(err)
	}
	installer := downloadTestInstaller(server, server.Client())
	if err := installer.download(context.Background(), "2.11.4", destination); err != nil {
		t.Fatalf("续传失败: %v", err)
	}
	result, _ := os.ReadFile(destination)
	if rangeHeader != "bytes=8-" || string(result) != string(payload) {
		t.Fatalf("续传结果不正确: range=%q payload=%q", rangeHeader, result)
	}
}

func TestCaddyDownloadRetriesAndKeepsExistingFileOnFailure(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		http.Error(writer, "mirror unavailable with internal details", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	destination := filepath.Join(t.TempDir(), "caddy.zip")
	if err := os.WriteFile(destination, []byte("old-runtime"), 0o600); err != nil {
		t.Fatal(err)
	}
	installer := downloadTestInstaller(server, server.Client())
	err := installer.download(context.Background(), "2.11.4", destination)
	var downloadErr *CaddyDownloadError
	if !errors.As(err, &downloadErr) {
		t.Fatalf("应返回结构化下载错误: %v", err)
	}
	if attempts.Load() != 3 || downloadErr.Attempt != 3 || downloadErr.HTTPStatus != http.StatusServiceUnavailable {
		t.Fatalf("重试信息不正确: attempts=%d error=%+v", attempts.Load(), downloadErr)
	}
	if !strings.Contains(err.Error(), server.URL) || !strings.Contains(err.Error(), "mirror unavailable") {
		t.Fatalf("错误缺少有效地址或响应: %v", err)
	}
	result, _ := os.ReadFile(destination)
	if string(result) != "old-runtime" {
		t.Fatalf("失败后覆盖了旧文件: %q", result)
	}
}

func downloadTestInstaller(server *httptest.Server, client HTTPDoer) *CaddyInstaller {
	return &CaddyInstaller{
		HTTPClient: client, GOOS: "windows", GOARCH: "amd64",
		DownloadURL: server.URL + "/caddy_{version}_{os}_{arch}.{ext}", RetryDelay: time.Millisecond,
	}
}
