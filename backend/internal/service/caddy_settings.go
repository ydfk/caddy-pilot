package service

import (
	"fmt"
	"net/url"
	"strings"

	"go-fiber-starter/internal/model/systemsetting"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	caddyVersionCheckURLKey     = "caddy.version_check_url"
	caddyDownloadURLKey         = "caddy.download_url"
	caddyChecksumURLKey         = "caddy.checksum_url"
	legacyCaddyPilotDownloadURL = "https://github.com/ydfk/caddy-pilot/releases/latest/download/caddy_{version}_{os}_{arch}.{ext}"
	legacyCaddyPilotChecksumURL = "https://github.com/ydfk/caddy-pilot/releases/latest/download/sha512sums.txt"
)

type CaddySettings struct {
	VersionCheckURL string
	DownloadURL     string
	ChecksumURL     string
}

func DefaultCaddySettings() CaddySettings {
	return CaddySettings{
		VersionCheckURL: CaddyLatestReleaseAPI,
		DownloadURL:     DefaultCaddyDownloadURL,
		ChecksumURL:     DefaultCaddyChecksumURL,
	}
}

func LoadCaddySettings(database *gorm.DB) (CaddySettings, error) {
	settings := DefaultCaddySettings()
	var records []systemsetting.Setting
	if err := database.Where("key IN ?", []string{
		caddyVersionCheckURLKey, caddyDownloadURLKey, caddyChecksumURLKey,
	}).Find(&records).Error; err != nil {
		return settings, fmt.Errorf("读取 Caddy 设置失败: %w", err)
	}
	for _, record := range records {
		switch record.Key {
		case caddyVersionCheckURLKey:
			settings.VersionCheckURL = record.Value
		case caddyDownloadURLKey:
			settings.DownloadURL = record.Value
		case caddyChecksumURLKey:
			settings.ChecksumURL = record.Value
		}
	}
	if isCaddyPilotDefaultSettings(settings) {
		settings = DefaultCaddySettings()
		if err := saveCaddySettings(database, settings); err != nil {
			return settings, fmt.Errorf("迁移 Caddy 默认更新源失败: %w", err)
		}
	}
	return settings, nil
}

func SaveCaddySettings(database *gorm.DB, settings CaddySettings) error {
	settings.VersionCheckURL = strings.TrimSpace(settings.VersionCheckURL)
	settings.DownloadURL = strings.TrimSpace(settings.DownloadURL)
	settings.ChecksumURL = strings.TrimSpace(settings.ChecksumURL)
	if err := validateCaddySettings(settings); err != nil {
		return err
	}
	return saveCaddySettings(database, settings)
}

func saveCaddySettings(database *gorm.DB, settings CaddySettings) error {
	records := []systemsetting.Setting{
		{Key: caddyVersionCheckURLKey, Value: settings.VersionCheckURL},
		{Key: caddyDownloadURLKey, Value: settings.DownloadURL},
		{Key: caddyChecksumURLKey, Value: settings.ChecksumURL},
	}
	return database.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&records).Error
}

func isCaddyPilotDefaultSettings(settings CaddySettings) bool {
	return settings.VersionCheckURL == CaddyPilotRuntimeManifestURL &&
		settings.DownloadURL == legacyCaddyPilotDownloadURL &&
		settings.ChecksumURL == legacyCaddyPilotChecksumURL
}

func validateCaddySettings(settings CaddySettings) error {
	if err := validateHTTPURL("版本校验地址", settings.VersionCheckURL, false); err != nil {
		return err
	}
	if err := validateHTTPURL("下载地址", settings.DownloadURL, false); err != nil {
		return err
	}
	return validateHTTPURL("SHA-512 清单地址", settings.ChecksumURL, true)
}

func validateHTTPURL(name, value string, optional bool) error {
	if value == "" && optional {
		return nil
	}
	if value == "" {
		return fmt.Errorf("%s不能为空", name)
	}
	normalized := strings.NewReplacer(
		"{version}", DefaultCaddyVersion, "{os}", "linux", "{arch}", "amd64", "{ext}", "tar.gz",
	).Replace(value)
	parsed, err := url.ParseRequestURI(normalized)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("%s必须是有效的 HTTP 或 HTTPS 地址", name)
	}
	return nil
}
