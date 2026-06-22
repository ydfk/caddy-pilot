package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"go-fiber-starter/internal/caddygen"
	"go-fiber-starter/internal/model/configversion"
	"go-fiber-starter/internal/model/proxysite"

	"gorm.io/gorm"
)

const (
	ConfigStatusDraft     = "draft"
	ConfigStatusPublished = "published"
	ConfigStatusFailed    = "failed"
	ConfigStatusRollback  = "rollback"
)

type ConfigService struct {
	DB    *gorm.DB
	Caddy CaddyAdmin
}

type ConfigChangeStatus struct {
	Dirty                  bool
	State                  string
	LatestVersion          uint
	LatestVersionID        uint
	ActiveVersion          uint
	RuntimeInSync          bool
	PersistentConfigInSync bool
	ErrorMessage           string
}

func NewConfigService(database *gorm.DB, caddy CaddyAdmin) *ConfigService {
	return &ConfigService{DB: database, Caddy: caddy}
}

func (service *ConfigService) Preview(ctx context.Context) ([]byte, error) {
	sites, err := service.enabledSites(ctx)
	if err != nil {
		return nil, err
	}
	if err := ResolveBasicAuthCredentials(ctx, service.DB, sites); err != nil {
		return nil, err
	}
	if err := ResolveCertificateProfiles(ctx, service.DB, sites); err != nil {
		return nil, err
	}
	return caddygen.Generate(sites)
}

func (service *ConfigService) ChangeStatus(ctx context.Context) (ConfigChangeStatus, error) {
	sites, err := service.enabledSites(ctx)
	if err != nil {
		return ConfigChangeStatus{}, err
	}
	current, err := json.Marshal(sites)
	if err != nil {
		return ConfigChangeStatus{}, fmt.Errorf("编码业务配置失败: %w", err)
	}

	var versions []configversion.ConfigVersion
	err = service.DB.WithContext(ctx).
		Where("status IN ?", []string{ConfigStatusPublished, ConfigStatusRollback}).
		Order("version DESC").Find(&versions).Error
	if err != nil {
		return ConfigChangeStatus{}, fmt.Errorf("读取最近发布配置失败: %w", err)
	}
	if len(versions) == 0 {
		return ConfigChangeStatus{Dirty: len(sites) > 0, State: "no_version"}, nil
	}
	latest := versions[0]
	status := ConfigChangeStatus{
		Dirty:           !sameJSON(current, []byte(latest.BusinessConfig)),
		LatestVersion:   latest.Version,
		LatestVersionID: latest.ID,
	}
	if service.Caddy == nil {
		status.State = "offline"
		status.ErrorMessage = "Caddy Admin API Client 未配置"
		return status, nil
	}
	runtimeConfig, runtimeErr := service.Caddy.GetConfig(ctx)
	if runtimeErr != nil {
		status.State = "offline"
		status.ErrorMessage = runtimeErr.Error()
		return status, nil
	}
	runtimeHash := normalizedJSONHash(runtimeConfig)
	for _, version := range versions {
		if runtimeHash == normalizedJSONHash([]byte(version.CaddyJSON)) {
			status.ActiveVersion = version.Version
			break
		}
	}
	status.RuntimeInSync = status.ActiveVersion > 0
	status.PersistentConfigInSync = runtimeMatchesActiveConfig(runtimeHash)
	switch {
	case !status.RuntimeInSync || !status.PersistentConfigInSync:
		status.State = "runtime_drift"
	case status.Dirty:
		status.State = "unpublished_changes"
	default:
		status.State = "in_sync"
	}

	return status, nil
}

func (service *ConfigService) Publish(ctx context.Context, reason string) (configversion.ConfigVersion, error) {
	sites, err := service.enabledSites(ctx)
	if err != nil {
		return configversion.ConfigVersion{}, err
	}
	businessConfig, err := json.Marshal(sites)
	if err != nil {
		return configversion.ConfigVersion{}, fmt.Errorf("编码业务配置失败: %w", err)
	}
	if err := ResolveBasicAuthCredentials(ctx, service.DB, sites); err != nil {
		return configversion.ConfigVersion{}, err
	}
	if err := ResolveCertificateProfiles(ctx, service.DB, sites); err != nil {
		return configversion.ConfigVersion{}, err
	}
	caddyJSON, generateErr := caddygen.Generate(sites)
	version, err := service.createAttempt(ctx, defaultReason(reason, "手动发布"), businessConfig, caddyJSON)
	if err != nil {
		return version, err
	}
	if generateErr != nil {
		return service.failAttempt(ctx, version, generateErr)
	}
	if !caddygen.HasManagementEntry(caddyJSON) {
		return service.failAttempt(ctx, version, errors.New("生成配置缺少 CaddyPilot 管理入口"))
	}
	if service.Caddy == nil {
		return service.failAttempt(ctx, version, errors.New("Caddy Admin API Client 未配置"))
	}
	if err := service.Caddy.LoadConfig(ctx, caddyJSON); err != nil {
		return service.failAttempt(ctx, version, err)
	}
	return service.completeAttempt(ctx, version, ConfigStatusPublished)
}

func (service *ConfigService) Rollback(ctx context.Context, id uint) (configversion.ConfigVersion, error) {
	var target configversion.ConfigVersion
	if err := service.DB.WithContext(ctx).First(&target, id).Error; err != nil {
		return configversion.ConfigVersion{}, err
	}
	protectedJSON, protectErr := caddygen.EnsureManagementEntry([]byte(target.CaddyJSON))
	version, err := service.createAttempt(
		ctx,
		fmt.Sprintf("回滚到版本 %d", target.Version),
		[]byte(target.BusinessConfig),
		protectedJSON,
	)
	if err != nil {
		return version, err
	}
	if protectErr != nil {
		return service.failAttempt(ctx, version, fmt.Errorf("修复历史配置管理入口失败: %w", protectErr))
	}
	if service.Caddy == nil {
		return service.failAttempt(ctx, version, errors.New("Caddy Admin API Client 未配置"))
	}
	if err := service.Caddy.LoadConfig(ctx, protectedJSON); err != nil {
		return service.failAttempt(ctx, version, err)
	}
	return service.completeAttempt(ctx, version, ConfigStatusRollback)
}

func ValidateCaddyJSON(payload []byte) error {
	if !json.Valid(payload) {
		return errors.New("Caddy JSON 格式无效")
	}
	if !caddygen.HasManagementEntry(payload) {
		return errors.New("Caddy JSON 缺少 CaddyPilot 管理入口")
	}
	return nil
}

func (service *ConfigService) enabledSites(ctx context.Context) ([]proxysite.ProxySite, error) {
	var sites []proxysite.ProxySite
	if err := service.DB.WithContext(ctx).Where("enabled = ?", true).Order("created_at ASC").Find(&sites).Error; err != nil {
		return nil, fmt.Errorf("读取启用站点失败: %w", err)
	}
	return sites, nil
}

func (service *ConfigService) createAttempt(ctx context.Context, reason string, businessConfig, caddyJSON []byte) (configversion.ConfigVersion, error) {
	versionNumber, err := service.nextVersion(ctx)
	if err != nil {
		return configversion.ConfigVersion{}, err
	}
	version := configversion.ConfigVersion{
		Version:        versionNumber,
		Reason:         reason,
		BusinessConfig: string(businessConfig),
		CaddyJSON:      string(caddyJSON),
		Status:         ConfigStatusDraft,
	}
	if err := service.DB.WithContext(ctx).Create(&version).Error; err != nil {
		return version, fmt.Errorf("保存配置版本失败: %w", err)
	}
	return version, nil
}

func (service *ConfigService) nextVersion(ctx context.Context) (uint, error) {
	var latest uint
	if err := service.DB.WithContext(ctx).Model(&configversion.ConfigVersion{}).Select("COALESCE(MAX(version), 0)").Scan(&latest).Error; err != nil {
		return 0, fmt.Errorf("生成配置版本号失败: %w", err)
	}
	return latest + 1, nil
}

func (service *ConfigService) failAttempt(ctx context.Context, version configversion.ConfigVersion, cause error) (configversion.ConfigVersion, error) {
	version.Status = ConfigStatusFailed
	version.ErrorMessage = cause.Error()
	if err := service.DB.WithContext(ctx).Model(&version).Updates(map[string]any{
		"status": version.Status, "error_message": version.ErrorMessage,
	}).Error; err != nil {
		return version, fmt.Errorf("%v；保存失败记录时出错: %w", cause, err)
	}
	return version, cause
}

func (service *ConfigService) completeAttempt(ctx context.Context, version configversion.ConfigVersion, status string) (configversion.ConfigVersion, error) {
	now := time.Now()
	version.Status = status
	version.PublishedAt = &now
	version.ErrorMessage = ""
	if err := service.DB.WithContext(ctx).Model(&version).Updates(map[string]any{
		"status": status, "published_at": now, "error_message": "",
	}).Error; err != nil {
		return version, fmt.Errorf("更新配置版本状态失败: %w", err)
	}
	return version, nil
}

func defaultReason(value, fallback string) string {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed
	}
	return fallback
}

func sameJSON(left, right []byte) bool {
	var leftValue, rightValue any
	if json.Unmarshal(left, &leftValue) != nil || json.Unmarshal(right, &rightValue) != nil {
		return bytes.Equal(bytes.TrimSpace(left), bytes.TrimSpace(right))
	}
	leftNormalized, _ := json.Marshal(leftValue)
	rightNormalized, _ := json.Marshal(rightValue)
	return bytes.Equal(leftNormalized, rightNormalized)
}

func normalizedJSONHash(payload []byte) [sha256.Size]byte {
	var value any
	if json.Unmarshal(payload, &value) == nil {
		payload, _ = json.Marshal(value)
	}
	return sha256.Sum256(bytes.TrimSpace(payload))
}

func runtimeMatchesActiveConfig(runtimeHash [sha256.Size]byte) bool {
	manager := ManagedCaddy()
	if manager == nil {
		return false
	}
	payload, err := os.ReadFile(manager.ActiveConfigPath())
	return err == nil && runtimeHash == normalizedJSONHash(payload)
}
