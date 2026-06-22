package dnsprovider

import (
	"context"
	"errors"
	"strings"

	model "go-fiber-starter/internal/model/dnsprovider"
	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/db"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func List(_ context.Context, _ *struct{}) (*DNSProviderListOutput, error) {
	var providers []model.DNSProvider
	if err := db.DB.Order("created_at DESC").Find(&providers).Error; err != nil {
		return nil, huma.Error500InternalServerError("查询 DNS Provider 失败")
	}
	output := &DNSProviderListOutput{Body: make([]DNSProviderResponse, 0, len(providers))}
	for _, provider := range providers {
		config, err := service.DecryptAliDNSConfig(provider.EncryptedConfig)
		if err != nil {
			return nil, huma.Error500InternalServerError("解密 DNS Provider 失败")
		}
		output.Body = append(output.Body, responseFromModel(provider, config.AccessKeyID, config.RegionID))
	}
	return output, nil
}

func Create(ctx context.Context, input *DNSProviderCreateInput) (*DNSProviderOutput, error) {
	if input.Body.ProviderType != "alidns" {
		return nil, huma.Error400BadRequest("当前只支持阿里云 DNS")
	}
	config := service.AliDNSConfig{
		AccessKeyID: strings.TrimSpace(input.Body.AccessKeyID), AccessKeySecret: input.Body.AccessKeySecret,
		RegionID: defaultRegion(input.Body.RegionID),
	}
	encrypted, err := service.EncryptAliDNSConfig(config)
	if err != nil {
		return nil, huma.Error500InternalServerError("加密 DNS Provider 失败")
	}
	provider := model.DNSProvider{
		Name: strings.TrimSpace(input.Body.Name), ProviderType: input.Body.ProviderType,
		EncryptedConfig: encrypted, Enabled: input.Body.Enabled,
	}
	if err := db.DB.Create(&provider).Error; err != nil {
		return nil, huma.Error500InternalServerError("创建 DNS Provider 失败")
	}
	if err := service.ApplyDNSProviderRuntime(ctx, db.DB); err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &DNSProviderOutput{Body: responseFromModel(provider, config.AccessKeyID, config.RegionID)}, nil
}

func Update(ctx context.Context, input *DNSProviderUpdateInput) (*DNSProviderOutput, error) {
	provider, err := find(input.ID)
	if err != nil {
		return nil, err
	}
	config, err := service.DecryptAliDNSConfig(provider.EncryptedConfig)
	if err != nil {
		return nil, huma.Error500InternalServerError("解密 DNS Provider 失败")
	}
	config.AccessKeyID = strings.TrimSpace(input.Body.AccessKeyID)
	config.RegionID = defaultRegion(input.Body.RegionID)
	if input.Body.AccessKeySecret != "" {
		config.AccessKeySecret = input.Body.AccessKeySecret
	}
	provider.Name = strings.TrimSpace(input.Body.Name)
	provider.Enabled = input.Body.Enabled
	provider.EncryptedConfig, err = service.EncryptAliDNSConfig(config)
	if err != nil {
		return nil, huma.Error500InternalServerError("加密 DNS Provider 失败")
	}
	if err := db.DB.Save(&provider).Error; err != nil {
		return nil, huma.Error500InternalServerError("更新 DNS Provider 失败")
	}
	if err := service.ApplyDNSProviderRuntime(ctx, db.DB); err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &DNSProviderOutput{Body: responseFromModel(provider, config.AccessKeyID, config.RegionID)}, nil
}

func Delete(ctx context.Context, input *DNSProviderIDInput) (*struct{}, error) {
	provider, err := find(input.ID)
	if err != nil {
		return nil, err
	}
	if err := db.DB.Delete(&provider).Error; err != nil {
		return nil, huma.Error500InternalServerError("删除 DNS Provider 失败")
	}
	if err := service.ApplyDNSProviderRuntime(ctx, db.DB); err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return nil, nil
}

func find(id string) (model.DNSProvider, error) {
	var provider model.DNSProvider
	if err := db.DB.First(&provider, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return provider, huma.Error404NotFound("DNS Provider 不存在")
		}
		return provider, huma.Error500InternalServerError("查询 DNS Provider 失败")
	}
	return provider, nil
}

func defaultRegion(value string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return "cn-hangzhou"
}
