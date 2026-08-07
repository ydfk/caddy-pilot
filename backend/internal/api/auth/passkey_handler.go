package auth

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"

	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var passkeyServiceCache struct {
	sync.Mutex
	database *gorm.DB
	config   config.PasskeyConfig
	service  *service.PasskeyService
}

func PasskeyStatus(ctx context.Context, _ *struct{}) (*PasskeyStatusOutput, error) {
	passkeys, err := currentPasskeyService()
	if err != nil {
		return &PasskeyStatusOutput{Body: PasskeyStatusResponse{
			Configured: false, ErrorMessage: err.Error(),
		}}, nil
	}
	available, err := passkeys.HasCredentials(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &PasskeyStatusOutput{Body: PasskeyStatusResponse{
		Configured: true, Available: available,
	}}, nil
}

func BeginPasskeyRegistration(ctx context.Context, input *PasskeyRegistrationOptionsInput) (*PasskeyCreationOptionsOutput, error) {
	userID, err := currentUserUUID(ctx)
	if err != nil {
		return nil, err
	}
	passkeys, err := currentPasskeyService()
	if err != nil {
		return nil, huma.Error503ServiceUnavailable(err.Error())
	}
	options, sessionID, err := passkeys.BeginRegistration(ctx, userID, input.Body.Name)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	return &PasskeyCreationOptionsOutput{Body: PasskeyCreationOptionsResponse{
		SessionID: sessionID, Options: *options,
	}}, nil
}

func FinishPasskeyRegistration(ctx context.Context, input *PasskeyVerifyInput) (*PasskeyOutput, error) {
	userID, err := currentUserUUID(ctx)
	if err != nil {
		return nil, err
	}
	if len(input.Body.Credential) == 0 {
		return nil, huma.Error400BadRequest("Passkey 凭据不能为空")
	}
	passkeys, err := currentPasskeyService()
	if err != nil {
		return nil, huma.Error503ServiceUnavailable(err.Error())
	}
	credential, err := passkeys.FinishRegistration(ctx, userID, input.Body.SessionID, input.Body.Credential)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	return &PasskeyOutput{Body: newPasskeyResponse(credential)}, nil
}

func BeginPasskeyLogin(ctx context.Context, _ *struct{}) (*PasskeyRequestOptionsOutput, error) {
	passkeys, err := currentPasskeyService()
	if err != nil {
		return nil, huma.Error503ServiceUnavailable(err.Error())
	}
	options, sessionID, err := passkeys.BeginLogin(ctx)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	return &PasskeyRequestOptionsOutput{Body: PasskeyRequestOptionsResponse{
		SessionID: sessionID, Options: *options,
	}}, nil
}

func FinishPasskeyLogin(ctx context.Context, input *PasskeyVerifyInput) (*TokenOutput, error) {
	if len(input.Body.Credential) == 0 {
		return nil, huma.Error400BadRequest("Passkey 凭据不能为空")
	}
	passkeys, err := currentPasskeyService()
	if err != nil {
		return nil, huma.Error503ServiceUnavailable(err.Error())
	}
	user, err := passkeys.FinishLogin(ctx, input.Body.SessionID, input.Body.Credential)
	if err != nil {
		return nil, huma.Error401Unauthorized(err.Error())
	}
	token, err := service.GenerateJWT(&user)
	if err != nil {
		return nil, huma.Error500InternalServerError("Token 生成失败")
	}
	return &TokenOutput{Body: TokenResponse{Token: token}}, nil
}

func ListPasskeys(ctx context.Context, _ *struct{}) (*PasskeyListOutput, error) {
	userID, err := currentUserUUID(ctx)
	if err != nil {
		return nil, err
	}
	passkeys, err := currentPasskeyService()
	if err != nil {
		return nil, huma.Error503ServiceUnavailable(err.Error())
	}
	credentials, err := passkeys.List(ctx, userID)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	items := make([]PasskeyResponse, 0, len(credentials))
	for _, credential := range credentials {
		items = append(items, newPasskeyResponse(credential))
	}
	return &PasskeyListOutput{Body: PasskeyListResponse{Items: items}}, nil
}

func RenamePasskey(ctx context.Context, input *PasskeyRenameInput) (*PasskeyOutput, error) {
	userID, err := currentUserUUID(ctx)
	if err != nil {
		return nil, err
	}
	credentialID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Passkey ID 无效")
	}
	passkeys, err := currentPasskeyService()
	if err != nil {
		return nil, huma.Error503ServiceUnavailable(err.Error())
	}
	credential, err := passkeys.Rename(ctx, userID, credentialID, input.Body.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, huma.Error404NotFound("Passkey 不存在")
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &PasskeyOutput{Body: newPasskeyResponse(credential)}, nil
}

func DeletePasskey(ctx context.Context, input *PasskeyIDInput) (*struct{}, error) {
	userID, err := currentUserUUID(ctx)
	if err != nil {
		return nil, err
	}
	credentialID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Passkey ID 无效")
	}
	passkeys, err := currentPasskeyService()
	if err != nil {
		return nil, huma.Error503ServiceUnavailable(err.Error())
	}
	if err := passkeys.Delete(ctx, userID, credentialID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, huma.Error404NotFound("Passkey 不存在")
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &struct{}{}, nil
}

func currentPasskeyService() (*service.PasskeyService, error) {
	passkeyServiceCache.Lock()
	defer passkeyServiceCache.Unlock()
	passkeyConfig := config.Current.Passkey
	if passkeyServiceCache.service != nil && passkeyServiceCache.database == db.DB &&
		reflect.DeepEqual(passkeyServiceCache.config, passkeyConfig) {
		return passkeyServiceCache.service, nil
	}
	passkeys, err := service.NewPasskeyService(db.DB, passkeyConfig)
	if err != nil {
		return nil, err
	}
	passkeyServiceCache.database = db.DB
	passkeyServiceCache.config = passkeyConfig
	passkeyServiceCache.service = passkeys
	return passkeys, nil
}

func currentUserUUID(ctx context.Context) (uuid.UUID, error) {
	value, ok := userIDFromContext(ctx)
	if !ok {
		return uuid.Nil, huma.Error401Unauthorized("认证信息无效")
	}
	userID, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return uuid.Nil, huma.Error401Unauthorized("用户 ID 无效")
	}
	return userID, nil
}
