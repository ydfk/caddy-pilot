package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"go-fiber-starter/internal/caddygen"

	"gopkg.in/natefinch/lumberjack.v2"
)

type CaddyManager struct {
	Installer  *CaddyInstaller
	Admin      *CaddyClient
	Runtime    CaddyRuntimeInfo
	configPath string
	command    *exec.Cmd
	process    *managedCaddyProcess
	caddyLog   io.WriteCloser
	lifecycle  chan error
	mu         sync.Mutex
}

type managedCaddyProcess struct {
	done     chan error
	expected atomic.Bool
	ready    atomic.Bool
}

var managedCaddy *CaddyManager

func NewCaddyManager() *CaddyManager {
	installer := NewCaddyInstaller()
	return &CaddyManager{
		Installer:  installer,
		Admin:      NewCaddyClient(),
		configPath: filepath.Join(installer.RuntimeDir, "caddy", "active.json"),
		lifecycle:  make(chan error, 1),
	}
}

func SetManagedCaddy(manager *CaddyManager) {
	managedCaddy = manager
}

func ManagedCaddy() *CaddyManager {
	return managedCaddy
}

func (manager *CaddyManager) Start(ctx context.Context, payload []byte) (CaddyRuntimeInfo, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	runtimeInfo, err := manager.Installer.Ensure(ctx)
	if err != nil {
		return CaddyRuntimeInfo{}, err
	}
	if err := manager.startLocked(ctx, runtimeInfo, payload); err != nil {
		return CaddyRuntimeInfo{}, err
	}
	return runtimeInfo, nil
}

func (manager *CaddyManager) ActiveConfigPath() string {
	return manager.configPath
}

func (manager *CaddyManager) GetConfig(ctx context.Context) ([]byte, error) {
	return manager.Admin.GetConfig(ctx)
}

func (manager *CaddyManager) GetStatus(ctx context.Context) error {
	return manager.Admin.GetStatus(ctx)
}

func (manager *CaddyManager) LoadConfig(ctx context.Context, payload []byte) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	protected, err := caddygen.EnsureManagementEntry(payload)
	if err != nil {
		return fmt.Errorf("保护管理入口失败: %w", err)
	}
	previous, err := manager.Admin.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("发布前读取 Caddy 配置失败: %w", err)
	}
	if err := manager.Admin.LoadConfig(ctx, protected); err != nil {
		return err
	}
	if err := writeFileAtomic(manager.configPath, protected, 0o600); err != nil {
		if restoreErr := manager.Admin.LoadConfig(ctx, previous); restoreErr != nil {
			return fmt.Errorf("持久化活动配置失败: %v；恢复旧配置失败: %w", err, restoreErr)
		}
		return fmt.Errorf("持久化活动配置失败，已恢复旧配置: %w", err)
	}
	return nil
}

func (manager *CaddyManager) Stop(ctx context.Context) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	return manager.stopLocked(ctx)
}

func (manager *CaddyManager) Restart(ctx context.Context) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	currentConfig, err := manager.Admin.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("重启前读取 Caddy 配置失败: %w", err)
	}
	protectedConfig, err := caddygen.EnsureManagementEntry(currentConfig)
	if err != nil {
		return fmt.Errorf("重启前保护管理入口失败: %w", err)
	}
	runtimeInfo := manager.Runtime
	if err := manager.stopLocked(ctx); err != nil {
		return err
	}
	return manager.startLocked(ctx, runtimeInfo, protectedConfig)
}

func (manager *CaddyManager) Update(ctx context.Context, version string, settings CaddySettings, report func(string, int64, int64)) (CaddyRuntimeInfo, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.Installer.DownloadURL = settings.DownloadURL
	manager.Installer.ChecksumURL = settings.ChecksumURL
	manager.Installer.Progress = func(downloaded, total int64) { report("downloading", downloaded, total) }
	manager.Installer.Stage = func(status string) { report(status, 0, 0) }
	defer func() {
		manager.Installer.Progress = nil
		manager.Installer.Stage = nil
	}()
	currentConfig, err := manager.Admin.GetConfig(ctx)
	if err != nil {
		return CaddyRuntimeInfo{}, fmt.Errorf("更新前读取 Caddy 配置失败: %w", err)
	}
	protectedConfig, err := caddygen.EnsureManagementEntry(currentConfig)
	if err != nil {
		return CaddyRuntimeInfo{}, fmt.Errorf("更新前保护管理入口失败: %w", err)
	}
	nextRuntime, err := manager.Installer.Install(ctx, version)
	if err != nil {
		return CaddyRuntimeInfo{}, err
	}
	report("restarting", 0, 0)
	return manager.switchRuntimeLocked(ctx, nextRuntime, protectedConfig)
}

func (manager *CaddyManager) UpdateUpload(ctx context.Context, uploadPath, filename string, report func(string, int64, int64)) (CaddyRuntimeInfo, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	defer os.Remove(uploadPath)
	manager.Installer.Stage = func(status string) { report(status, 0, 0) }
	defer func() { manager.Installer.Stage = nil }()
	currentConfig, err := manager.Admin.GetConfig(ctx)
	if err != nil {
		return CaddyRuntimeInfo{}, fmt.Errorf("更新前读取 Caddy 配置失败: %w", err)
	}
	protectedConfig, err := caddygen.EnsureManagementEntry(currentConfig)
	if err != nil {
		return CaddyRuntimeInfo{}, fmt.Errorf("更新前保护管理入口失败: %w", err)
	}
	nextRuntime, err := manager.Installer.InstallUpload(ctx, uploadPath, filename)
	if err != nil {
		return CaddyRuntimeInfo{}, err
	}
	report("restarting", 0, 0)
	return manager.switchRuntimeLocked(ctx, nextRuntime, protectedConfig)
}

func (manager *CaddyManager) switchRuntimeLocked(ctx context.Context, nextRuntime CaddyRuntimeInfo, protectedConfig []byte) (CaddyRuntimeInfo, error) {
	previousRuntime := manager.Runtime
	if err := manager.stopLocked(ctx); err != nil {
		return CaddyRuntimeInfo{}, err
	}
	if err := manager.startLocked(ctx, nextRuntime, protectedConfig); err != nil {
		restoreContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if restoreErr := manager.startLocked(restoreContext, previousRuntime, protectedConfig); restoreErr != nil {
			return CaddyRuntimeInfo{}, fmt.Errorf("启动新 Caddy 失败: %v；恢复旧版本失败: %w", err, restoreErr)
		}
		return CaddyRuntimeInfo{}, fmt.Errorf("启动新 Caddy 失败，已恢复旧版本: %w", err)
	}
	if err := manager.Installer.Select(nextRuntime.Version); err != nil {
		return CaddyRuntimeInfo{}, fmt.Errorf("保存 Caddy 当前版本失败: %w", err)
	}
	return nextRuntime, nil
}

func (manager *CaddyManager) Done() <-chan error {
	return manager.lifecycle
}

func (manager *CaddyManager) startLocked(ctx context.Context, runtimeInfo CaddyRuntimeInfo, payload []byte) error {
	if err := os.MkdirAll(filepath.Dir(manager.configPath), 0o755); err != nil {
		return fmt.Errorf("创建 Caddy 配置目录失败: %w", err)
	}
	if err := writeFileAtomic(manager.configPath, payload, 0o600); err != nil {
		return fmt.Errorf("写入 Caddy 初始配置失败: %w", err)
	}
	process := &managedCaddyProcess{done: make(chan error, 1)}
	manager.process = process
	manager.command = exec.Command(runtimeInfo.BinaryPath, "run", "--config", manager.configPath)
	logWriter, err := newCaddyLogWriter()
	if err != nil {
		return err
	}
	manager.caddyLog = logWriter
	manager.command.Stdout = io.MultiWriter(os.Stdout, logWriter)
	manager.command.Stderr = io.MultiWriter(os.Stderr, logWriter)
	dataDir := environmentValue("CADDY_DATA_DIR", filepath.Join(manager.Installer.RuntimeDir, "caddy-data"))
	manager.command.Env = append(os.Environ(), "XDG_DATA_HOME="+dataDir)
	if err := manager.command.Start(); err != nil {
		_ = manager.caddyLog.Close()
		manager.caddyLog = nil
		return fmt.Errorf("启动 Caddy 失败: %w", err)
	}
	go func(command *exec.Cmd) {
		err := command.Wait()
		process.done <- err
		unexpected := !process.expected.Load() && process.ready.Load()
		if unexpected {
			select {
			case manager.lifecycle <- fmt.Errorf("Caddy 进程意外退出: %w", err):
			default:
			}
		}
	}(manager.command)
	if err := manager.waitUntilReady(ctx, process.done); err != nil {
		process.expected.Store(true)
		_ = manager.command.Process.Kill()
		return err
	}
	manager.Runtime = runtimeInfo
	process.ready.Store(true)
	_ = os.Setenv("CADDY_BINARY", runtimeInfo.BinaryPath)
	return nil
}

func writeFileAtomic(path string, payload []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".config-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(mode); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(payload); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return replaceFile(temporaryPath, path)
}

func (manager *CaddyManager) stopLocked(ctx context.Context) error {
	if manager.command == nil || manager.command.Process == nil {
		return nil
	}
	process := manager.process
	process.expected.Store(true)
	process.ready.Store(false)
	request, requestErr := http.NewRequestWithContext(ctx, http.MethodPost, manager.Admin.AdminAPI+"/stop", bytes.NewReader(nil))
	var response *http.Response
	if requestErr == nil {
		response, requestErr = manager.Admin.HTTPClient.Do(request)
	}
	if response != nil {
		response.Body.Close()
	}
	select {
	case <-process.done:
		manager.command = nil
		manager.closeCaddyLog()
		return nil
	case <-ctx.Done():
		_ = manager.command.Process.Kill()
		manager.closeCaddyLog()
		return ctx.Err()
	case <-time.After(5 * time.Second):
		_ = manager.command.Process.Kill()
		manager.command = nil
		manager.closeCaddyLog()
		if requestErr != nil {
			return fmt.Errorf("停止 Caddy 失败: %w", requestErr)
		}
		return fmt.Errorf("停止 Caddy 超时")
	}
}

func newCaddyLogWriter() (io.WriteCloser, error) {
	logDir := environmentValue("CADDYPILOT_LOG_DIR", filepath.Join("data", "logs"))
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建 Caddy 日志目录失败: %w", err)
	}
	return &lumberjack.Logger{
		Filename: filepath.Join(logDir, "caddy.log"), MaxSize: 10, MaxBackups: 3, MaxAge: 28, Compress: true,
	}, nil
}

func (manager *CaddyManager) closeCaddyLog() {
	if manager.caddyLog != nil {
		_ = manager.caddyLog.Close()
		manager.caddyLog = nil
	}
}

func (manager *CaddyManager) waitUntilReady(ctx context.Context, processDone <-chan error) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()
	for {
		if err := manager.Admin.GetStatus(ctx); err == nil {
			return nil
		}
		select {
		case err := <-processDone:
			return fmt.Errorf("Caddy 启动后退出: %w", err)
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return fmt.Errorf("等待 Caddy Admin API 就绪超时")
		case <-ticker.C:
		}
	}
}
