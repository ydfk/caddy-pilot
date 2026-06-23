package service

import (
	"testing"

	"go-fiber-starter/internal/model/systemsetting"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCaddySettingsPersistAndLoad(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&systemsetting.Setting{}); err != nil {
		t.Fatal(err)
	}
	want := CaddySettings{
		VersionCheckURL: "https://mirror.example/caddy/latest.json",
		DownloadURL:     "https://mirror.example/caddy/{version}/{os}/{arch}.{ext}",
		ChecksumURL:     "https://mirror.example/caddy/{version}/sha512.txt",
	}
	if err := SaveCaddySettings(database, want); err != nil {
		t.Fatalf("保存 Caddy 设置失败: %v", err)
	}
	got, err := LoadCaddySettings(database)
	if err != nil {
		t.Fatalf("读取 Caddy 设置失败: %v", err)
	}
	if got != want {
		t.Fatalf("Caddy 设置不一致: %+v", got)
	}
}

func TestCaddySettingsRejectInvalidURL(t *testing.T) {
	settings := DefaultCaddySettings()
	settings.DownloadURL = "file:///tmp/caddy"
	if err := validateCaddySettings(settings); err == nil {
		t.Fatal("应拒绝非 HTTP 下载地址")
	}
}

func TestCaddySettingsMigratesLegacyDefaults(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&systemsetting.Setting{}); err != nil {
		t.Fatal(err)
	}
	legacy := []systemsetting.Setting{
		{Key: caddyVersionCheckURLKey, Value: CaddyLatestReleaseAPI},
		{Key: caddyDownloadURLKey, Value: legacyCaddyDownloadURL},
		{Key: caddyChecksumURLKey, Value: ""},
	}
	if err := database.Create(&legacy).Error; err != nil {
		t.Fatal(err)
	}
	settings, err := LoadCaddySettings(database)
	if err != nil {
		t.Fatal(err)
	}
	if settings != DefaultCaddySettings() {
		t.Fatalf("旧默认源未迁移: %+v", settings)
	}
}
