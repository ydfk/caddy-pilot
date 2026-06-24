package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func ValidateCaddyfile(ctx context.Context, payload []byte) error {
	_, err := AdaptCaddyfile(ctx, payload)
	return err
}

func ValidateCaddyConfig(ctx context.Context, payload []byte) error {
	file, err := os.CreateTemp("", "caddypilot-*.json")
	if err != nil {
		return fmt.Errorf("创建 Caddy 配置校验文件失败: %w", err)
	}
	path := file.Name()
	defer os.Remove(path)
	if _, err := file.Write(payload); err != nil {
		file.Close()
		return fmt.Errorf("写入 Caddy 配置校验文件失败: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("关闭 Caddy 配置校验文件失败: %w", err)
	}

	validationContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	command := exec.CommandContext(validationContext, caddyBinary(), "validate", "--config", path)
	output, err := command.CombinedOutput()
	if err != nil {
		message := caddyErrorMessage(output)
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("Caddy 配置校验失败: %s", message)
	}
	return nil
}

func caddyErrorMessage(output []byte) string {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for index := len(lines) - 1; index >= 0; index-- {
		line := strings.TrimSpace(lines[index])
		if strings.HasPrefix(line, "Error:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Error:"))
		}
	}
	return strings.TrimSpace(string(output))
}

func AdaptCaddyfile(ctx context.Context, payload []byte) ([]byte, error) {
	file, err := os.CreateTemp("", "caddypilot-*.Caddyfile")
	if err != nil {
		return nil, fmt.Errorf("创建 Caddyfile 校验文件失败: %w", err)
	}
	path := file.Name()
	defer os.Remove(path)
	if _, err := file.Write(payload); err != nil {
		file.Close()
		return nil, fmt.Errorf("写入 Caddyfile 校验文件失败: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("关闭 Caddyfile 校验文件失败: %w", err)
	}
	validationContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	command := exec.CommandContext(validationContext, caddyBinary(), "adapt", "--adapter", "caddyfile", "--config", path, "--pretty")
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("Caddyfile 语法校验失败: %s", message)
	}
	return output, nil
}

func caddyBinary() string {
	binary := environmentValue("CADDY_BINARY", "caddy")
	if manager := ManagedCaddy(); manager != nil && manager.Runtime.BinaryPath != "" {
		return manager.Runtime.BinaryPath
	}
	return binary
}
