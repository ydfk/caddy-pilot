package service

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type CaddyRuntimeInfo struct {
	BinaryPath string
	Version    string
}

func CheckCaddyRuntime(ctx context.Context) (CaddyRuntimeInfo, error) {
	return NewCaddyInstaller().Ensure(ctx)
}

func inspectCaddyBinary(ctx context.Context, path string) (CaddyRuntimeInfo, error) {
	checkContext, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	output, err := exec.CommandContext(checkContext, path, "version").CombinedOutput()
	if err != nil {
		return CaddyRuntimeInfo{}, fmt.Errorf("执行 Caddy 版本检查失败: %w", err)
	}
	fields := strings.Fields(string(output))
	if len(fields) == 0 {
		return CaddyRuntimeInfo{}, fmt.Errorf("Caddy 版本输出为空")
	}
	return CaddyRuntimeInfo{BinaryPath: path, Version: normalizeVersion(fields[0])}, nil
}
