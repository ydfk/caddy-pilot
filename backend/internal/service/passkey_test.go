package service

import (
	"context"
	"testing"

	passkeymodel "go-fiber-starter/internal/model/passkey"
	usermodel "go-fiber-starter/internal/model/user"
	"go-fiber-starter/pkg/config"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPasskeyServiceRegistrationOptionsAndCredentialManagement(t *testing.T) {
	previousConfig := config.Current
	config.Current.Jwt.Secret = "passkey-test-secret"
	t.Cleanup(func() { config.Current = previousConfig })

	database, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := database.AutoMigrate(&usermodel.User{}, &passkeymodel.Credential{}); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}
	user := usermodel.User{Username: "admin", Password: "unused"}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	passkeys, err := NewPasskeyService(database, config.PasskeyConfig{
		RPID: "localhost", DisplayName: "CaddyPilot Test", Origins: []string{"http://localhost:8080"},
	})
	if err != nil {
		t.Fatalf("创建 Passkey 服务失败: %v", err)
	}
	creation, sessionID, err := passkeys.BeginRegistration(context.Background(), user.Id, "MacBook")
	if err != nil {
		t.Fatalf("创建注册挑战失败: %v", err)
	}
	if sessionID == "" || creation.Response.RelyingParty.ID != "localhost" {
		t.Fatalf("注册挑战无效: session=%q rp=%q", sessionID, creation.Response.RelyingParty.ID)
	}

	record, err := passkeys.newCredentialRecord(user.Id, "MacBook", &webauthn.Credential{ID: []byte("credential-id")})
	if err != nil {
		t.Fatalf("创建凭据记录失败: %v", err)
	}
	if record.CredentialIDHash == "" || record.EncryptedCredential == "" {
		t.Fatal("凭据 ID 应散列且凭据内容应加密")
	}
	if err := database.Create(&record).Error; err != nil {
		t.Fatalf("保存凭据失败: %v", err)
	}

	hasCredentials, err := passkeys.HasCredentials(context.Background())
	if err != nil || !hasCredentials {
		t.Fatalf("Passkey 状态错误: available=%v err=%v", hasCredentials, err)
	}
	renamed, err := passkeys.Rename(context.Background(), user.Id, record.Id, "安全密钥")
	if err != nil || renamed.Name != "安全密钥" {
		t.Fatalf("重命名 Passkey 失败: %+v err=%v", renamed, err)
	}
	if err := passkeys.Delete(context.Background(), user.Id, record.Id); err != nil {
		t.Fatalf("删除 Passkey 失败: %v", err)
	}
}

func TestPasskeySessionIsSingleUse(t *testing.T) {
	service := &PasskeyService{sessions: make(map[string]passkeySession)}
	sessionID, err := service.storeSession(passkeySession{Kind: "login"})
	if err != nil {
		t.Fatalf("保存会话失败: %v", err)
	}
	if _, err := service.takeSession(" "+sessionID+" ", "login", uuid.Nil); err != nil {
		t.Fatalf("消费会话失败: %v", err)
	}
	if _, err := service.takeSession(sessionID, "login", uuid.Nil); err == nil {
		t.Fatal("Passkey 会话不应重复使用")
	}
}

func TestPasskeyServiceRejectsInvalidConfiguration(t *testing.T) {
	previousConfig := config.Current
	config.Current.Jwt.Secret = "passkey-test-secret"
	t.Cleanup(func() { config.Current = previousConfig })
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if _, err := NewPasskeyService(database, config.PasskeyConfig{}); err == nil {
		t.Fatal("缺少 RP ID 和 Origin 时应拒绝初始化")
	}
}
