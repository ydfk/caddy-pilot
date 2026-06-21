package service

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"go-fiber-starter/internal/caddygen"
)

type CaddyManager struct {
	Installer  *CaddyInstaller
	Admin      *CaddyClient
	Runtime    CaddyRuntimeInfo
	configPath string
	command    *exec.Cmd
	process    *managedCaddyProcess
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
		configPath: filepath.Join(installer.RuntimeDir, "caddy", "initial.json"),
		lifecycle:  make(chan error, 1),
	}
}

func SetManagedCaddy(manager *CaddyManager) {
	managedCaddy = manager
}

func ManagedCaddy() *CaddyManager {
	return managedCaddy
}

func (manager *CaddyManager) Start(ctx context.Context) (CaddyRuntimeInfo, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	runtimeInfo, err := manager.Installer.Ensure(ctx)
	if err != nil {
		return CaddyRuntimeInfo{}, err
	}
	payload, err := caddygen.Generate(nil)
	if err != nil {
		return CaddyRuntimeInfo{}, fmt.Errorf("生成 Caddy 初始配置失败: %w", err)
	}
	if err := manager.startLocked(ctx, runtimeInfo, payload); err != nil {
		return CaddyRuntimeInfo{}, err
	}
	return runtimeInfo, nil
}

func (manager *CaddyManager) Stop(ctx context.Context) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	return manager.stopLocked(ctx)
}

func (manager *CaddyManager) Update(ctx context.Context, version string) (CaddyRuntimeInfo, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
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
	if err := os.WriteFile(manager.configPath, payload, 0o600); err != nil {
		return fmt.Errorf("写入 Caddy 初始配置失败: %w", err)
	}
	process := &managedCaddyProcess{done: make(chan error, 1)}
	manager.process = process
	manager.command = exec.Command(runtimeInfo.BinaryPath, "run", "--config", manager.configPath)
	manager.command.Stdout = os.Stdout
	manager.command.Stderr = os.Stderr
	dataDir := environmentValue("CADDY_DATA_DIR", filepath.Join(manager.Installer.RuntimeDir, "caddy-data"))
	manager.command.Env = append(os.Environ(), "XDG_DATA_HOME="+dataDir)
	if err := manager.command.Start(); err != nil {
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
		return nil
	case <-ctx.Done():
		_ = manager.command.Process.Kill()
		return ctx.Err()
	case <-time.After(5 * time.Second):
		_ = manager.command.Process.Kill()
		manager.command = nil
		if requestErr != nil {
			return fmt.Errorf("停止 Caddy 失败: %w", requestErr)
		}
		return fmt.Errorf("停止 Caddy 超时")
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
