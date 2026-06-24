package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"go-fiber-starter/internal/caddygen"
	"go-fiber-starter/internal/model/configversion"
	"go-fiber-starter/internal/model/proxysite"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeCaddyAdmin struct {
	loaded  []byte
	loadErr error
}

func (fake *fakeCaddyAdmin) GetConfig(context.Context) ([]byte, error) {
	return fake.loaded, nil
}

func (fake *fakeCaddyAdmin) LoadConfig(_ context.Context, payload []byte) error {
	fake.loaded = append([]byte(nil), payload...)
	return fake.loadErr
}

func (fake *fakeCaddyAdmin) GetStatus(context.Context) error {
	return fake.loadErr
}

func TestConfigServicePublishAndRollbackKeepManagementEntry(t *testing.T) {
	database := configServiceTestDB(t)
	createEnabledTestSite(t, database)
	fake := &fakeCaddyAdmin{}
	service := NewConfigService(database, fake)

	published, err := service.Publish(context.Background(), "首次发布")
	if err != nil {
		t.Fatalf("发布配置失败: %v", err)
	}
	if published.Status != ConfigStatusPublished || published.Caddyfile == "" || !caddygen.HasManagementEntry(fake.loaded) {
		t.Fatalf("发布状态或管理入口不正确: %+v", published)
	}

	legacy := configversion.ConfigVersion{
		Version: 100, Reason: "旧配置", BusinessConfig: "[]",
		CaddyJSON: `{"apps":{"http":{"servers":{}}}}`, Caddyfile: ":8080 {}", Status: ConfigStatusPublished,
	}
	if err := database.Create(&legacy).Error; err != nil {
		t.Fatalf("创建历史配置失败: %v", err)
	}
	rolledBack, err := service.Rollback(context.Background(), legacy.ID)
	if err != nil {
		t.Fatalf("回滚配置失败: %v", err)
	}
	if rolledBack.Status != ConfigStatusRollback || rolledBack.Caddyfile != legacy.Caddyfile || !caddygen.HasManagementEntry(fake.loaded) {
		t.Fatalf("回滚状态或管理入口不正确: %+v", rolledBack)
	}
}

func TestConfigServicePublishFailureIsRecorded(t *testing.T) {
	database := configServiceTestDB(t)
	createEnabledTestSite(t, database)
	fake := &fakeCaddyAdmin{loadErr: errors.New("连接失败")}
	service := NewConfigService(database, fake)

	version, err := service.Publish(context.Background(), "失败发布")
	if err == nil || version.Status != ConfigStatusFailed {
		t.Fatalf("发布失败未正确返回: %+v, %v", version, err)
	}
	var stored configversion.ConfigVersion
	if err := database.First(&stored, version.ID).Error; err != nil {
		t.Fatalf("读取失败版本失败: %v", err)
	}
	if stored.Status != ConfigStatusFailed || stored.ErrorMessage == "" {
		t.Fatalf("失败版本未留痕: %+v", stored)
	}
}

func TestConfigServiceRollbackFailureIsRecorded(t *testing.T) {
	database := configServiceTestDB(t)
	payload, err := caddygen.Generate(nil)
	if err != nil {
		t.Fatalf("生成历史配置失败: %v", err)
	}
	target := configversion.ConfigVersion{
		Version: 1, Reason: "历史版本", BusinessConfig: "[]",
		CaddyJSON: string(payload), Status: ConfigStatusPublished,
	}
	if err := database.Create(&target).Error; err != nil {
		t.Fatalf("创建历史版本失败: %v", err)
	}
	service := NewConfigService(database, &fakeCaddyAdmin{loadErr: errors.New("回滚连接失败")})

	failed, err := service.Rollback(context.Background(), target.ID)
	if err == nil || failed.Status != ConfigStatusFailed {
		t.Fatalf("回滚失败未正确返回: %+v, %v", failed, err)
	}
	var stored configversion.ConfigVersion
	if err := database.First(&stored, failed.ID).Error; err != nil {
		t.Fatalf("读取失败回滚记录失败: %v", err)
	}
	if stored.Status != ConfigStatusFailed || stored.ErrorMessage == "" {
		t.Fatalf("回滚失败未留痕: %+v", stored)
	}
}

func TestConfigServiceChangeStatus(t *testing.T) {
	database := configServiceTestDB(t)
	service := NewConfigService(database, &fakeCaddyAdmin{})

	status, err := service.ChangeStatus(context.Background())
	if err != nil || status.Dirty {
		t.Fatalf("空配置不应标记为未发布: %+v, %v", status, err)
	}

	createEnabledTestSite(t, database)
	status, err = service.ChangeStatus(context.Background())
	if err != nil || !status.Dirty {
		t.Fatalf("新增启用站点后应标记为未发布: %+v, %v", status, err)
	}

	published, err := service.Publish(context.Background(), "状态测试")
	if err != nil {
		t.Fatalf("发布配置失败: %v", err)
	}
	status, err = service.ChangeStatus(context.Background())
	if err != nil || status.Dirty || status.LatestVersion != published.Version {
		t.Fatalf("发布后不应存在未发布变更: %+v, %v", status, err)
	}

	var stored configversion.ConfigVersion
	if err := database.First(&stored, published.ID).Error; err != nil {
		t.Fatal(err)
	}
	var historical []map[string]any
	if err := json.Unmarshal([]byte(stored.BusinessConfig), &historical); err != nil {
		t.Fatal(err)
	}
	delete(historical[0], "config_mode")
	legacyPayload, _ := json.Marshal(historical)
	if err := database.Model(&stored).Update("business_config", string(legacyPayload)).Error; err != nil {
		t.Fatal(err)
	}
	status, err = service.ChangeStatus(context.Background())
	if err != nil || status.Dirty {
		t.Fatalf("旧版本缺少新增默认字段时不应误报未发布: %+v, %v", status, err)
	}

	if err := database.Model(&proxysite.ProxySite{}).Where("enabled = ?", true).Update("description", "已修改").Error; err != nil {
		t.Fatalf("修改站点失败: %v", err)
	}
	status, err = service.ChangeStatus(context.Background())
	if err != nil || !status.Dirty {
		t.Fatalf("修改启用站点后应标记为未发布: %+v, %v", status, err)
	}
}

func configServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := database.AutoMigrate(&proxysite.ProxySite{}, &configversion.ConfigVersion{}); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}
	return database
}

func createEnabledTestSite(t *testing.T, database *gorm.DB) {
	t.Helper()
	encode := func(value any) string {
		payload, _ := json.Marshal(value)
		return string(payload)
	}
	site := proxysite.ProxySite{
		Name: "测试站点", Domains: encode([]string{"example.com"}),
		Upstreams:      encode([]string{"127.0.0.1:3000"}),
		RequestHeaders: encode(map[string]string{}), ResponseHeaders: encode(map[string]string{}),
		BasicAuthUsers: encode(map[string]string{}), AllowedIPs: encode([]string{}), Enabled: true,
	}
	if err := database.Create(&site).Error; err != nil {
		t.Fatalf("创建测试站点失败: %v", err)
	}
}
