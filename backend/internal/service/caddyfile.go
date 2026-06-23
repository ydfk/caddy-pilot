package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func ValidateCaddyfile(ctx context.Context, payload []byte) error {
	file, err := os.CreateTemp("", "caddypilot-*.Caddyfile")
	if err != nil {
		return fmt.Errorf("创建 Caddyfile 校验文件失败: %w", err)
	}
	path := file.Name()
	defer os.Remove(path)
	if _, err := file.Write(payload); err != nil {
		file.Close()
		return fmt.Errorf("写入 Caddyfile 校验文件失败: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("关闭 Caddyfile 校验文件失败: %w", err)
	}
	binary := environmentValue("CADDY_BINARY", "caddy")
	if manager := ManagedCaddy(); manager != nil && manager.Runtime.BinaryPath != "" {
		binary = manager.Runtime.BinaryPath
	}
	validationContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	output, err := exec.CommandContext(validationContext, binary, "adapt", "--adapter", "caddyfile", "--config", path).CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("Caddyfile 语法校验失败: %s", message)
	}
	return nil
}
