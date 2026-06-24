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
	binary := environmentValue("CADDY_BINARY", "caddy")
	if manager := ManagedCaddy(); manager != nil && manager.Runtime.BinaryPath != "" {
		binary = manager.Runtime.BinaryPath
	}
	validationContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	command := exec.CommandContext(validationContext, binary, "adapt", "--adapter", "caddyfile", "--config", path, "--pretty")
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
