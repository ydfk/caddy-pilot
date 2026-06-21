package service

import (
	"context"
	"encoding/json"
	"fmt"

	"go-fiber-starter/internal/model/basicauth"
	"go-fiber-starter/internal/model/proxysite"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ResolveBasicAuthCredentials(ctx context.Context, database *gorm.DB, sites []proxysite.ProxySite) error {
	for index := range sites {
		if !sites[index].BasicAuthEnabled {
			continue
		}
		var ids []uuid.UUID
		if err := json.Unmarshal([]byte(defaultJSON(sites[index].BasicAuthCredentialIDs, "[]")), &ids); err != nil {
			return fmt.Errorf("解析站点 %s 的密码引用失败: %w", sites[index].Name, err)
		}
		if len(ids) == 0 {
			continue
		}
		var credentials []basicauth.Credential
		if err := database.WithContext(ctx).Where("id IN ?", ids).Find(&credentials).Error; err != nil {
			return fmt.Errorf("读取站点 %s 的密码条目失败: %w", sites[index].Name, err)
		}
		if len(credentials) != len(ids) {
			return fmt.Errorf("站点 %s 引用了不存在的密码条目", sites[index].Name)
		}
		users := make(map[string]string, len(credentials))
		for _, credential := range credentials {
			if _, exists := users[credential.Username]; exists {
				return fmt.Errorf("站点 %s 选择了重复用户名 %s", sites[index].Name, credential.Username)
			}
			users[credential.Username] = credential.PasswordHash
		}
		payload, _ := json.Marshal(users)
		sites[index].BasicAuthUsers = string(payload)
	}
	return nil
}

func defaultJSON(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
