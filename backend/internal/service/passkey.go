package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	passkeymodel "go-fiber-starter/internal/model/passkey"
	usermodel "go-fiber-starter/internal/model/user"
	"go-fiber-starter/pkg/config"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	passkeySessionLifetime = 5 * time.Minute
	passkeyMaxSessions     = 1024
)

type PasskeyService struct {
	database *gorm.DB
	webAuthn *webauthn.WebAuthn
	box      *SecretBox
	mu       sync.Mutex
	sessions map[string]passkeySession
}

type passkeySession struct {
	Kind      string
	UserID    uuid.UUID
	Name      string
	Data      webauthn.SessionData
	ExpiresAt time.Time
}

type PasskeyUser struct {
	User        usermodel.User
	Credentials []webauthn.Credential
}

func NewPasskeyService(database *gorm.DB, passkeyConfig config.PasskeyConfig) (*PasskeyService, error) {
	if database == nil {
		return nil, fmt.Errorf("Passkey 数据库尚未初始化")
	}
	box, err := NewSecretBox()
	if err != nil {
		return nil, fmt.Errorf("初始化 Passkey 加密失败: %w", err)
	}
	webAuthn, err := webauthn.New(&webauthn.Config{
		RPID:          strings.TrimSpace(passkeyConfig.RPID),
		RPDisplayName: strings.TrimSpace(passkeyConfig.DisplayName),
		RPOrigins:     normalizedOrigins(passkeyConfig.Origins),
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementRequired,
			UserVerification: protocol.VerificationRequired,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Passkey 配置无效，请检查 RP ID 和允许的 Origin: %w", err)
	}
	return &PasskeyService{
		database: database,
		webAuthn: webAuthn,
		box:      box,
		sessions: make(map[string]passkeySession),
	}, nil
}

func normalizedOrigins(origins []string) []string {
	result := make([]string, 0, len(origins))
	for _, origin := range origins {
		if value := strings.TrimSpace(origin); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func (user *PasskeyUser) WebAuthnID() []byte {
	id := user.User.Id
	return id[:]
}

func (user *PasskeyUser) WebAuthnName() string {
	return user.User.Username
}

func (user *PasskeyUser) WebAuthnDisplayName() string {
	return user.User.Username
}

func (user *PasskeyUser) WebAuthnCredentials() []webauthn.Credential {
	return user.Credentials
}

func (service *PasskeyService) HasCredentials(ctx context.Context) (bool, error) {
	var count int64
	if err := service.database.WithContext(ctx).Model(&passkeymodel.Credential{}).Count(&count).Error; err != nil {
		return false, fmt.Errorf("检查 Passkey 状态失败: %w", err)
	}
	return count > 0, nil
}

func (service *PasskeyService) List(ctx context.Context, userID uuid.UUID) ([]passkeymodel.Credential, error) {
	var credentials []passkeymodel.Credential
	err := service.database.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&credentials).Error
	if err != nil {
		return nil, fmt.Errorf("读取 Passkey 失败: %w", err)
	}
	return credentials, nil
}

func (service *PasskeyService) BeginRegistration(ctx context.Context, userID uuid.UUID, name string) (*protocol.CredentialCreation, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, "", fmt.Errorf("Passkey 名称不能为空")
	}
	user, err := service.loadUser(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	creation, session, err := service.webAuthn.BeginRegistration(
		user,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithExclusions(webauthn.Credentials(user.Credentials).CredentialDescriptors()),
	)
	if err != nil {
		return nil, "", fmt.Errorf("创建 Passkey 注册挑战失败: %w", err)
	}
	sessionID, err := service.storeSession(passkeySession{
		Kind: "registration", UserID: userID, Name: name, Data: *session,
	})
	if err != nil {
		return nil, "", err
	}
	return creation, sessionID, nil
}

func (service *PasskeyService) FinishRegistration(ctx context.Context, userID uuid.UUID, sessionID string, response []byte) (passkeymodel.Credential, error) {
	session, err := service.takeSession(sessionID, "registration", userID)
	if err != nil {
		return passkeymodel.Credential{}, err
	}
	user, err := service.loadUser(ctx, userID)
	if err != nil {
		return passkeymodel.Credential{}, err
	}
	parsed, err := protocol.ParseCredentialCreationResponseBytes(response)
	if err != nil {
		return passkeymodel.Credential{}, fmt.Errorf("解析 Passkey 注册响应失败: %w", err)
	}
	credential, err := service.webAuthn.CreateCredential(user, session.Data, parsed)
	if err != nil {
		return passkeymodel.Credential{}, fmt.Errorf("验证 Passkey 注册响应失败: %w", err)
	}
	record, err := service.newCredentialRecord(userID, session.Name, credential)
	if err != nil {
		return passkeymodel.Credential{}, err
	}
	if err := service.database.WithContext(ctx).Create(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return passkeymodel.Credential{}, fmt.Errorf("此 Passkey 已经登记")
		}
		return passkeymodel.Credential{}, fmt.Errorf("保存 Passkey 失败: %w", err)
	}
	return record, nil
}

func (service *PasskeyService) BeginLogin(ctx context.Context) (*protocol.CredentialAssertion, string, error) {
	hasCredentials, err := service.HasCredentials(ctx)
	if err != nil {
		return nil, "", err
	}
	if !hasCredentials {
		return nil, "", fmt.Errorf("尚未登记 Passkey，请先使用密码登录并添加")
	}
	assertion, session, err := service.webAuthn.BeginDiscoverableLogin(
		webauthn.WithUserVerification(protocol.VerificationRequired),
	)
	if err != nil {
		return nil, "", fmt.Errorf("创建 Passkey 登录挑战失败: %w", err)
	}
	sessionID, err := service.storeSession(passkeySession{Kind: "login", Data: *session})
	if err != nil {
		return nil, "", err
	}
	return assertion, sessionID, nil
}

func (service *PasskeyService) FinishLogin(ctx context.Context, sessionID string, response []byte) (usermodel.User, error) {
	session, err := service.takeSession(sessionID, "login", uuid.Nil)
	if err != nil {
		return usermodel.User{}, err
	}
	parsed, err := protocol.ParseCredentialRequestResponseBytes(response)
	if err != nil {
		return usermodel.User{}, fmt.Errorf("解析 Passkey 登录响应失败: %w", err)
	}
	validatedUser, credential, err := service.webAuthn.ValidatePasskeyLogin(
		func(rawID, userHandle []byte) (webauthn.User, error) {
			return service.loadDiscoverableUser(ctx, rawID, userHandle)
		},
		session.Data,
		parsed,
	)
	if err != nil {
		return usermodel.User{}, fmt.Errorf("Passkey 验证失败: %w", err)
	}
	user, ok := validatedUser.(*PasskeyUser)
	if !ok {
		return usermodel.User{}, fmt.Errorf("Passkey 用户类型无效")
	}
	if err := service.updateCredential(ctx, user.User.Id, credential); err != nil {
		return usermodel.User{}, err
	}
	return user.User, nil
}

func (service *PasskeyService) Rename(ctx context.Context, userID, credentialID uuid.UUID, name string) (passkeymodel.Credential, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return passkeymodel.Credential{}, fmt.Errorf("Passkey 名称不能为空")
	}
	var credential passkeymodel.Credential
	result := service.database.WithContext(ctx).Model(&credential).
		Where("id = ? AND user_id = ?", credentialID, userID).
		Update("name", name)
	if result.Error != nil {
		return credential, fmt.Errorf("重命名 Passkey 失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return credential, gorm.ErrRecordNotFound
	}
	if err := service.database.WithContext(ctx).First(&credential, "id = ?", credentialID).Error; err != nil {
		return credential, fmt.Errorf("读取 Passkey 失败: %w", err)
	}
	return credential, nil
}

func (service *PasskeyService) Delete(ctx context.Context, userID, credentialID uuid.UUID) error {
	result := service.database.WithContext(ctx).
		Where("id = ? AND user_id = ?", credentialID, userID).
		Delete(&passkeymodel.Credential{})
	if result.Error != nil {
		return fmt.Errorf("删除 Passkey 失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (service *PasskeyService) loadUser(ctx context.Context, userID uuid.UUID) (*PasskeyUser, error) {
	var user usermodel.User
	if err := service.database.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	records, err := service.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	credentials := make([]webauthn.Credential, 0, len(records))
	for _, record := range records {
		credential, decryptErr := service.decryptCredential(record.EncryptedCredential)
		if decryptErr != nil {
			return nil, fmt.Errorf("解密 Passkey %s 失败: %w", record.Name, decryptErr)
		}
		credentials = append(credentials, credential)
	}
	return &PasskeyUser{User: user, Credentials: credentials}, nil
}

func (service *PasskeyService) loadDiscoverableUser(ctx context.Context, rawID, userHandle []byte) (webauthn.User, error) {
	userID, err := uuid.FromBytes(userHandle)
	if err != nil {
		return nil, fmt.Errorf("Passkey 用户标识无效")
	}
	credentialIDHash := hashCredentialID(rawID)
	var count int64
	if err := service.database.WithContext(ctx).Model(&passkeymodel.Credential{}).
		Where("user_id = ? AND credential_id_hash = ?", userID, credentialIDHash).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count != 1 {
		return nil, fmt.Errorf("Passkey 不属于此用户")
	}
	return service.loadUser(ctx, userID)
}

func (service *PasskeyService) newCredentialRecord(userID uuid.UUID, name string, credential *webauthn.Credential) (passkeymodel.Credential, error) {
	encrypted, err := service.encryptCredential(credential)
	if err != nil {
		return passkeymodel.Credential{}, fmt.Errorf("加密 Passkey 失败: %w", err)
	}
	return passkeymodel.Credential{
		UserID:              userID,
		Name:                name,
		CredentialIDHash:    hashCredentialID(credential.ID),
		EncryptedCredential: encrypted,
	}, nil
}

func (service *PasskeyService) updateCredential(ctx context.Context, userID uuid.UUID, credential *webauthn.Credential) error {
	encrypted, err := service.encryptCredential(credential)
	if err != nil {
		return fmt.Errorf("加密 Passkey 失败: %w", err)
	}
	now := time.Now()
	result := service.database.WithContext(ctx).Model(&passkeymodel.Credential{}).
		Where("user_id = ? AND credential_id_hash = ?", userID, hashCredentialID(credential.ID)).
		Updates(map[string]any{"encrypted_credential": encrypted, "last_used_at": &now})
	if result.Error != nil {
		return fmt.Errorf("更新 Passkey 状态失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("找不到已验证的 Passkey")
	}
	return nil
}

func (service *PasskeyService) encryptCredential(credential *webauthn.Credential) (string, error) {
	payload, err := json.Marshal(credential)
	if err != nil {
		return "", err
	}
	return service.box.Encrypt(payload)
}

func (service *PasskeyService) decryptCredential(value string) (webauthn.Credential, error) {
	payload, err := service.box.Decrypt(value)
	if err != nil {
		return webauthn.Credential{}, err
	}
	var credential webauthn.Credential
	if err := json.Unmarshal(payload, &credential); err != nil {
		return webauthn.Credential{}, err
	}
	return credential, nil
}

func (service *PasskeyService) storeSession(session passkeySession) (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("生成 Passkey 会话失败: %w", err)
	}
	sessionID := base64.RawURLEncoding.EncodeToString(buffer)
	session.ExpiresAt = time.Now().Add(passkeySessionLifetime)
	service.mu.Lock()
	defer service.mu.Unlock()
	service.removeExpiredSessionsLocked()
	if len(service.sessions) >= passkeyMaxSessions {
		service.removeOldestSessionLocked()
	}
	service.sessions[sessionID] = session
	return sessionID, nil
}

func (service *PasskeyService) takeSession(sessionID, kind string, userID uuid.UUID) (passkeySession, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	service.removeExpiredSessionsLocked()
	normalizedSessionID := strings.TrimSpace(sessionID)
	session, ok := service.sessions[normalizedSessionID]
	if !ok {
		return passkeySession{}, fmt.Errorf("Passkey 会话不存在或已过期，请重试")
	}
	delete(service.sessions, normalizedSessionID)
	if session.Kind != kind || (userID != uuid.Nil && session.UserID != userID) {
		return passkeySession{}, fmt.Errorf("Passkey 会话与当前操作不匹配")
	}
	return session, nil
}

func hashCredentialID(credentialID []byte) string {
	hash := sha256.Sum256(credentialID)
	return hex.EncodeToString(hash[:])
}

func (service *PasskeyService) removeExpiredSessionsLocked() {
	now := time.Now()
	for sessionID, session := range service.sessions {
		if !session.ExpiresAt.After(now) {
			delete(service.sessions, sessionID)
		}
	}
}

func (service *PasskeyService) removeOldestSessionLocked() {
	var oldestID string
	var oldestExpiry time.Time
	for sessionID, session := range service.sessions {
		if oldestID == "" || session.ExpiresAt.Before(oldestExpiry) {
			oldestID = sessionID
			oldestExpiry = session.ExpiresAt
		}
	}
	if oldestID != "" {
		delete(service.sessions, oldestID)
	}
}
