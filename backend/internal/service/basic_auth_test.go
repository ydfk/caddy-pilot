package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"go-fiber-starter/internal/model/basicauth"
	"go-fiber-starter/internal/model/proxysite"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestResolveBasicAuthCredentials(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := database.AutoMigrate(&basicauth.Credential{}); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}
	credential := basicauth.Credential{Name: "运维", Username: "admin", PasswordHash: "$2a$10$hash"}
	if err := database.Create(&credential).Error; err != nil {
		t.Fatalf("创建密码条目失败: %v", err)
	}
	ids, _ := json.Marshal([]uuid.UUID{credential.Id})
	sites := []proxysite.ProxySite{{Name: "example.com", BasicAuthEnabled: true, BasicAuthCredentialIDs: string(ids)}}
	if err := ResolveBasicAuthCredentials(context.Background(), database, sites); err != nil {
		t.Fatalf("解析密码条目失败: %v", err)
	}
	if !strings.Contains(sites[0].BasicAuthUsers, credential.PasswordHash) {
		t.Fatalf("站点未写入密码哈希: %s", sites[0].BasicAuthUsers)
	}
}
