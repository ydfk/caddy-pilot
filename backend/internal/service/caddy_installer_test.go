package service

import (
	"archive/zip"
	"context"
	"crypto/sha512"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCaddyInstallerBuildsPlatformDownloadURL(t *testing.T) {
	installer := &CaddyInstaller{
		GOOS:        "windows",
		GOARCH:      "amd64",
		DownloadURL: DefaultCaddyDownloadURL,
		ChecksumURL: DefaultCaddyChecksumURL,
	}
	want := "https://caddyserver.com/api/download?os=windows&arch=amd64&p=github.com/caddy-dns/alidns&v=2.10.0"
	if got := installer.downloadURL("2.10.0"); got != want {
		t.Fatalf("下载地址不正确: %s", got)
	}
	checksumWant := ""
	if got := installer.checksumURL("2.10.0"); got != checksumWant {
		t.Fatalf("校验和地址不正确: %s", got)
	}
}

func TestCaddyInstallerVerifiesChecksum(t *testing.T) {
	payload := []byte("caddy-archive")
	hash := sha512.Sum512(payload)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(writer, "%x  caddy_2.10.0_windows_amd64.zip\n", hash)
	}))
	defer server.Close()
	archivePath := filepath.Join(t.TempDir(), "caddy.zip")
	if err := os.WriteFile(archivePath, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	installer := &CaddyInstaller{
		HTTPClient:  server.Client(),
		GOOS:        "windows",
		GOARCH:      "amd64",
		ChecksumURL: server.URL,
	}
	if err := installer.verifyChecksum(context.Background(), "2.10.0", archivePath); err != nil {
		t.Fatalf("校验 Caddy 下载失败: %v", err)
	}
}

func TestExtractCaddyZip(t *testing.T) {
	directory := t.TempDir()
	archivePath := filepath.Join(directory, "caddy.zip")
	destination := filepath.Join(directory, "caddy.exe")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(archiveFile)
	entry, err := archive.Create("caddy.exe")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("managed-caddy")); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archiveFile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := extractCaddyZip(archivePath, destination, "caddy.exe"); err != nil {
		t.Fatalf("解压 Caddy 失败: %v", err)
	}
	payload, err := os.ReadFile(destination)
	if err != nil || string(payload) != "managed-caddy" {
		t.Fatalf("解压结果不正确: %q, %v", payload, err)
	}
}

func TestReplaceFileOverwritesExistingDestination(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "source")
	destination := filepath.Join(directory, "destination")
	if err := os.WriteFile(source, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := replaceFile(source, destination); err != nil {
		t.Fatalf("替换文件失败: %v", err)
	}
	payload, err := os.ReadFile(destination)
	if err != nil || string(payload) != "new" {
		t.Fatalf("替换结果不正确: %q, %v", payload, err)
	}
}
