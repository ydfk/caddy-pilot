package service

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go-fiber-starter/internal/caddygen"
	"go-fiber-starter/internal/model/configversion"

	"gorm.io/gorm"
)

func StartupCaddyConfig(ctx context.Context, database *gorm.DB, activePath string) ([]byte, error) {
	if payload, err := os.ReadFile(activePath); err == nil {
		if protected, protectErr := caddygen.EnsureManagementEntry(payload); protectErr == nil {
			return protected, nil
		}
	}

	var latest configversion.ConfigVersion
	err := database.WithContext(ctx).
		Where("status IN ?", []string{ConfigStatusPublished, ConfigStatusRollback}).
		Order("version DESC").First(&latest).Error
	if err == nil {
		protected, protectErr := caddygen.EnsureManagementEntry([]byte(latest.CaddyJSON))
		if protectErr != nil {
			return nil, fmt.Errorf("恢复最近成功配置失败: %w", protectErr)
		}
		return protected, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("读取最近成功配置失败: %w", err)
	}
	return caddygen.Generate(nil)
}
