package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go-fiber-starter/internal/model/dnsprovider"

	"gorm.io/gorm"
)

type AliDNSConfig struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	RegionID        string `json:"region_id,omitempty"`
}

func EncryptAliDNSConfig(value AliDNSConfig) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	box, err := NewSecretBox()
	if err != nil {
		return "", err
	}
	return box.Encrypt(payload)
}

func DecryptAliDNSConfig(value string) (AliDNSConfig, error) {
	box, err := NewSecretBox()
	if err != nil {
		return AliDNSConfig{}, err
	}
	payload, err := box.Decrypt(value)
	if err != nil {
		return AliDNSConfig{}, err
	}
	var result AliDNSConfig
	if err := json.Unmarshal(payload, &result); err != nil {
		return AliDNSConfig{}, err
	}
	return result, nil
}

func LoadDNSProviderEnvironment(ctx context.Context, database *gorm.DB) error {
	var providers []dnsprovider.DNSProvider
	if err := database.WithContext(ctx).Where("enabled = ?", true).Find(&providers).Error; err != nil {
		return fmt.Errorf("读取 DNS Provider 失败: %w", err)
	}
	for _, provider := range providers {
		if provider.ProviderType != "alidns" {
			continue
		}
		config, err := DecryptAliDNSConfig(provider.EncryptedConfig)
		if err != nil {
			return fmt.Errorf("解密 DNS Provider %s 失败: %w", provider.Name, err)
		}
		idName, secretName, regionName := dnsprovider.EnvNames(provider.Id)
		if err := os.Setenv(idName, config.AccessKeyID); err != nil {
			return err
		}
		if err := os.Setenv(secretName, config.AccessKeySecret); err != nil {
			return err
		}
		if err := os.Setenv(regionName, config.RegionID); err != nil {
			return err
		}
	}
	return nil
}

func ApplyDNSProviderRuntime(ctx context.Context, database *gorm.DB) error {
	if err := LoadDNSProviderEnvironment(ctx, database); err != nil {
		return err
	}
	manager := ManagedCaddy()
	if manager == nil {
		return nil
	}
	if err := manager.Restart(ctx); err != nil {
		return fmt.Errorf("重启托管 Caddy 失败: %w", err)
	}
	return nil
}
