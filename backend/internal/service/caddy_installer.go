package service

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	DefaultCaddyVersion     = "2.11.4"
	DefaultCaddyDownloadURL = "https://caddyserver.com/api/download?os={os}&arch={arch}&p=github.com/caddy-dns/alidns&v={version}"
	DefaultCaddyChecksumURL = ""
	MaxCaddyUploadSize      = 128 << 20
)

type CaddyInstaller struct {
	RuntimeDir  string
	HTTPClient  HTTPDoer
	GOOS        string
	GOARCH      string
	DownloadURL string
	ChecksumURL string
	Progress    func(CaddyUpdateProgress)
	RetryDelay  time.Duration
}

func NewCaddyInstaller() *CaddyInstaller {
	return &CaddyInstaller{
		RuntimeDir:  environmentValue("CADDYPILOT_RUNTIME_DIR", filepath.Join("data", "runtime")),
		HTTPClient:  newCaddyHTTPClient(),
		GOOS:        runtime.GOOS,
		GOARCH:      runtime.GOARCH,
		DownloadURL: DefaultCaddyDownloadURL,
		ChecksumURL: DefaultCaddyChecksumURL,
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
			return installer.importRuntime(ctx, runtimeInfo)
		}
	}
	return installer.Install(ctx, environmentValue("CADDY_VERSION", DefaultCaddyVersion))
}

func (installer *CaddyInstaller) importRuntime(ctx context.Context, runtimeInfo CaddyRuntimeInfo) (CaddyRuntimeInfo, error) {
	target := installer.versionedBinary(runtimeInfo.Version)
	if samePath(runtimeInfo.BinaryPath, target) {
		if err := installer.Select(runtimeInfo.Version); err != nil {
			return CaddyRuntimeInfo{}, err
		}
		return runtimeInfo, nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return CaddyRuntimeInfo{}, fmt.Errorf("创建 Caddy 运行目录失败: %w", err)
	}
	source, err := os.Open(runtimeInfo.BinaryPath)
	if err != nil {
		return CaddyRuntimeInfo{}, fmt.Errorf("读取内置 Caddy 失败: %w", err)
	}
	defer source.Close()
	if err := writeExecutable(target, source); err != nil {
		return CaddyRuntimeInfo{}, err
	}
	imported, err := inspectCaddyBinary(ctx, target)
	if err != nil {
		return CaddyRuntimeInfo{}, err
	}
	if err := installer.Select(imported.Version); err != nil {
		return CaddyRuntimeInfo{}, err
	}
	return imported, nil
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
	installer.report(CaddyUpdateProgress{Stage: "downloading"})
	if err := installer.download(ctx, version, archivePath); err != nil {
		return CaddyRuntimeInfo{}, err
	}
	defer os.Remove(archivePath)
	if strings.TrimSpace(installer.ChecksumURL) != "" {
		installer.report(CaddyUpdateProgress{Stage: "verifying"})
		if err := installer.verifyChecksum(ctx, version, archivePath); err != nil {
			return CaddyRuntimeInfo{}, err
		}
	}
	if installer.isDirectBinaryDownload() {
		installer.report(CaddyUpdateProgress{Stage: "installing"})
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
	installer.report(CaddyUpdateProgress{Stage: "installing"})
	if err := installer.extractBinary(archivePath, target); err != nil {
		return CaddyRuntimeInfo{}, err
	}
	return inspectCaddyBinary(ctx, target)
}

func (installer *CaddyInstaller) isDirectBinaryDownload() bool {
	return strings.Contains(installer.DownloadURL, "caddyserver.com/api/download")
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

func (installer *CaddyInstaller) SaveUpload(source io.Reader) (string, error) {
	directory := filepath.Join(installer.RuntimeDir, "caddy", "uploads")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", fmt.Errorf("创建 Caddy 上传目录失败: %w", err)
	}
	file, err := os.CreateTemp(directory, "upload-*")
	if err != nil {
		return "", fmt.Errorf("创建 Caddy 上传文件失败: %w", err)
	}
	path := file.Name()
	written, copyErr := io.Copy(file, io.LimitReader(source, MaxCaddyUploadSize+1))
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil || written > MaxCaddyUploadSize {
		_ = os.Remove(path)
		if written > MaxCaddyUploadSize {
			return "", fmt.Errorf("Caddy 上传文件超过 128 MiB 限制")
		}
		if copyErr != nil {
			return "", fmt.Errorf("保存 Caddy 上传文件失败: %w", copyErr)
		}
		return "", fmt.Errorf("关闭 Caddy 上传文件失败: %w", closeErr)
	}
	return path, nil
}

func (installer *CaddyInstaller) InstallUpload(ctx context.Context, uploadPath, filename string) (CaddyRuntimeInfo, error) {
	installer.report(CaddyUpdateProgress{Stage: "verifying"})
	directory, err := os.MkdirTemp(filepath.Join(installer.RuntimeDir, "caddy"), ".install-")
	if err != nil {
		return CaddyRuntimeInfo{}, fmt.Errorf("创建 Caddy 安装目录失败: %w", err)
	}
	defer os.RemoveAll(directory)
	candidate := filepath.Join(directory, installer.binaryName())
	lowerName := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lowerName, ".zip"):
		err = extractCaddyZip(uploadPath, candidate, installer.binaryName())
	case strings.HasSuffix(lowerName, ".tar.gz"), strings.HasSuffix(lowerName, ".tgz"):
		err = extractCaddyTarGzip(uploadPath, candidate, installer.binaryName())
	default:
		var source *os.File
		source, err = os.Open(uploadPath)
		if err == nil {
			defer source.Close()
			err = writeExecutable(candidate, source)
		}
	}
	if err != nil {
		return CaddyRuntimeInfo{}, fmt.Errorf("解包 Caddy 上传文件失败: %w", err)
	}
	runtimeInfo, err := inspectCaddyBinary(ctx, candidate)
	if err != nil {
		return CaddyRuntimeInfo{}, fmt.Errorf("校验上传的 Caddy 失败: %w", err)
	}
	target := installer.versionedBinary(runtimeInfo.Version)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return CaddyRuntimeInfo{}, err
	}
	source, err := os.Open(candidate)
	if err != nil {
		return CaddyRuntimeInfo{}, err
	}
	defer source.Close()
	if err := writeExecutable(target, source); err != nil {
		return CaddyRuntimeInfo{}, err
	}
	return inspectCaddyBinary(ctx, target)
}

func (installer *CaddyInstaller) report(progress CaddyUpdateProgress) {
	if installer.Progress != nil {
		installer.Progress(progress)
	}
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
	if err := replaceFile(temporary, destination); err != nil {
		return fmt.Errorf("安装 Caddy 可执行文件失败: %w", err)
	}
	return nil
}

func replaceFile(source, destination string) error {
	if err := os.Rename(source, destination); err == nil {
		return nil
	}
	backup := destination + ".bak"
	_ = os.Remove(backup)
	hadDestination := false
	if err := os.Rename(destination, backup); err == nil {
		hadDestination = true
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(source, destination); err != nil {
		if hadDestination {
			_ = os.Rename(backup, destination)
		}
		return err
	}
	if hadDestination {
		_ = os.Remove(backup)
	}
	return nil
}

func samePath(left, right string) bool {
	leftPath, leftErr := filepath.Abs(left)
	rightPath, rightErr := filepath.Abs(right)
	return leftErr == nil && rightErr == nil && strings.EqualFold(filepath.Clean(leftPath), filepath.Clean(rightPath))
}
