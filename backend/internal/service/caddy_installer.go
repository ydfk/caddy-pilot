package service

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	DefaultCaddyVersion     = "2.10.0"
	DefaultCaddyDownloadURL = "https://caddyserver.com/api/download?os={os}&arch={arch}&p=github.com/caddy-dns/alidns&v={version}"
	DefaultCaddyChecksumURL = ""
)

type CaddyInstaller struct {
	RuntimeDir  string
	HTTPClient  *http.Client
	GOOS        string
	GOARCH      string
	DownloadURL string
	ChecksumURL string
}

func NewCaddyInstaller() *CaddyInstaller {
	return &CaddyInstaller{
		RuntimeDir:  environmentValue("CADDYPILOT_RUNTIME_DIR", filepath.Join("data", "runtime")),
		HTTPClient:  &http.Client{Timeout: 5 * time.Minute},
		GOOS:        runtime.GOOS,
		GOARCH:      runtime.GOARCH,
		DownloadURL: environmentValue("CADDY_DOWNLOAD_URL", DefaultCaddyDownloadURL),
		ChecksumURL: environmentValue("CADDY_CHECKSUM_URL", DefaultCaddyChecksumURL),
	}
}

func (installer *CaddyInstaller) Ensure(ctx context.Context) (CaddyRuntimeInfo, error) {
	if selected, ok := installer.selectedBinary(); ok {
		if runtimeInfo, err := inspectCaddyBinary(ctx, selected); err == nil {
			return runtimeInfo, nil
		}
	}
	for _, candidate := range installer.binaryCandidates() {
		path, err := exec.LookPath(candidate)
		if err != nil {
			continue
		}
		if runtimeInfo, err := inspectCaddyBinary(ctx, path); err == nil {
			return runtimeInfo, nil
		}
	}
	return installer.Install(ctx, environmentValue("CADDY_VERSION", DefaultCaddyVersion))
}

func (installer *CaddyInstaller) Install(ctx context.Context, version string) (CaddyRuntimeInfo, error) {
	version = normalizeVersion(version)
	if version == "" {
		return CaddyRuntimeInfo{}, fmt.Errorf("Caddy 目标版本不能为空")
	}
	target := installer.versionedBinary(version)
	if _, err := os.Stat(target); err == nil {
		if runtimeInfo, inspectErr := inspectCaddyBinary(ctx, target); inspectErr == nil {
			return runtimeInfo, nil
		}
		if err := os.Remove(target); err != nil {
			return CaddyRuntimeInfo{}, fmt.Errorf("替换缺少阿里云 DNS 模块的 Caddy 失败: %w", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return CaddyRuntimeInfo{}, fmt.Errorf("创建 Caddy 运行目录失败: %w", err)
	}

	archivePath := filepath.Join(filepath.Dir(target), "caddy-download."+installer.archiveExtension())
	if err := installer.download(ctx, version, archivePath); err != nil {
		return CaddyRuntimeInfo{}, err
	}
	defer os.Remove(archivePath)
	if installer.ChecksumURL != "" {
		if err := installer.verifyChecksum(ctx, version, archivePath); err != nil {
			return CaddyRuntimeInfo{}, err
		}
	}
	if installer.isDirectBinaryDownload() {
		source, err := os.Open(archivePath)
		if err != nil {
			return CaddyRuntimeInfo{}, err
		}
		defer source.Close()
		if err := writeExecutable(target, source); err != nil {
			return CaddyRuntimeInfo{}, err
		}
		return inspectCaddyBinary(ctx, target)
	}
	if err := installer.extractBinary(archivePath, target); err != nil {
		return CaddyRuntimeInfo{}, err
	}
	return inspectCaddyBinary(ctx, target)
}

func (installer *CaddyInstaller) isDirectBinaryDownload() bool {
	return strings.Contains(installer.DownloadURL, "caddyserver.com/api/download")
}

func (installer *CaddyInstaller) verifyChecksum(ctx context.Context, version, archivePath string) error {
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
		return fmt.Errorf("下载 Caddy 校验和失败: 校验服务返回 %d", response.StatusCode)
	}
	wantedName := installer.archiveName(version)
	wantedHash := ""
	scanner := bufio.NewScanner(io.LimitReader(response.Body, 1<<20))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && strings.TrimPrefix(fields[len(fields)-1], "*") == wantedName {
			wantedHash = fields[0]
			break
		}
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
	actualHash := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actualHash, wantedHash) {
		return fmt.Errorf("Caddy 下载文件校验失败")
	}
	return nil
}

func (installer *CaddyInstaller) Select(version string) error {
	version = normalizeVersion(version)
	if _, err := os.Stat(installer.versionedBinary(version)); err != nil {
		return fmt.Errorf("选择的 Caddy %s 不存在: %w", version, err)
	}
	if err := os.MkdirAll(installer.RuntimeDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(installer.selectionPath(), []byte(version), 0o644)
}

func (installer *CaddyInstaller) binaryCandidates() []string {
	candidates := make([]string, 0, 3)
	if configured := strings.TrimSpace(os.Getenv("CADDY_BINARY")); configured != "" {
		candidates = append(candidates, configured)
	}
	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), installer.binaryName()))
	}
	return append(candidates, installer.binaryName())
}

func (installer *CaddyInstaller) selectedBinary() (string, bool) {
	payload, err := os.ReadFile(installer.selectionPath())
	if err != nil {
		return "", false
	}
	path := installer.versionedBinary(strings.TrimSpace(string(payload)))
	_, err = os.Stat(path)
	return path, err == nil
}

func (installer *CaddyInstaller) download(ctx context.Context, version, destination string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, installer.downloadURL(version), nil)
	if err != nil {
		return fmt.Errorf("创建 Caddy 下载请求失败: %w", err)
	}
	request.Header.Set("User-Agent", "CaddyPilot")
	response, err := installer.HTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("下载 Caddy %s 失败: %w", version, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("下载 Caddy %s 失败: 下载服务返回 %d", version, response.StatusCode)
	}

	file, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("创建 Caddy 下载文件失败: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, io.LimitReader(response.Body, 128<<20)); err != nil {
		return fmt.Errorf("保存 Caddy 下载文件失败: %w", err)
	}
	return nil
}

func (installer *CaddyInstaller) extractBinary(archivePath, destination string) error {
	if installer.GOOS == "windows" {
		return extractCaddyZip(archivePath, destination, installer.binaryName())
	}
	return extractCaddyTarGzip(archivePath, destination, installer.binaryName())
}

func (installer *CaddyInstaller) downloadURL(version string) string {
	return installer.expandURL(installer.DownloadURL, version)
}

func (installer *CaddyInstaller) checksumURL(version string) string {
	return installer.expandURL(installer.ChecksumURL, version)
}

func (installer *CaddyInstaller) expandURL(template, version string) string {
	replacer := strings.NewReplacer(
		"{version}", version,
		"{os}", installer.GOOS,
		"{arch}", installer.GOARCH,
		"{ext}", installer.archiveExtension(),
	)
	return replacer.Replace(template)
}

func (installer *CaddyInstaller) archiveName(version string) string {
	return fmt.Sprintf("caddy_%s_%s_%s.%s", version, installer.GOOS, installer.GOARCH, installer.archiveExtension())
}

func (installer *CaddyInstaller) versionedBinary(version string) string {
	return filepath.Join(installer.RuntimeDir, "caddy", normalizeVersion(version), installer.binaryName())
}

func (installer *CaddyInstaller) selectionPath() string {
	return filepath.Join(installer.RuntimeDir, "caddy", "current-version")
}

func (installer *CaddyInstaller) binaryName() string {
	if installer.GOOS == "windows" {
		return "caddy.exe"
	}
	return "caddy"
}

func (installer *CaddyInstaller) archiveExtension() string {
	if installer.GOOS == "windows" {
		return "zip"
	}
	return "tar.gz"
}

func extractCaddyZip(archivePath, destination, binaryName string) error {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("打开 Caddy zip 失败: %w", err)
	}
	defer archive.Close()
	for _, item := range archive.File {
		if filepath.Base(item.Name) != binaryName {
			continue
		}
		source, err := item.Open()
		if err != nil {
			return err
		}
		defer source.Close()
		return writeExecutable(destination, source)
	}
	return fmt.Errorf("Caddy zip 中缺少 %s", binaryName)
}

func extractCaddyTarGzip(archivePath, destination, binaryName string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("打开 Caddy tar.gz 失败: %w", err)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if filepath.Base(header.Name) == binaryName && header.Typeflag == tar.TypeReg {
			return writeExecutable(destination, tarReader)
		}
	}
	return fmt.Errorf("Caddy tar.gz 中缺少 %s", binaryName)
}

func writeExecutable(destination string, source io.Reader) error {
	temporary := destination + ".tmp"
	file, err := os.OpenFile(temporary, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return fmt.Errorf("创建 Caddy 可执行文件失败: %w", err)
	}
	if _, err := io.Copy(file, source); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporary, destination); err != nil {
		return fmt.Errorf("安装 Caddy 可执行文件失败: %w", err)
	}
	return nil
}
