package auth

import (
	"encoding/json"
	"time"

	passkeymodel "go-fiber-starter/internal/model/passkey"
	model "go-fiber-starter/internal/model/user"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
)

type CredentialsInput struct {
	Body Credentials
}

type Credentials struct {
	Username string `json:"username" minLength:"1" maxLength:"64" example:"admin" doc:"用户名"`
	Password string `json:"password" minLength:"6" maxLength:"72" example:"change-me" doc:"密码"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id" doc:"用户 ID"`
	Username  string    `json:"username" doc:"用户名"`
	CreatedAt time.Time `json:"createdAt" doc:"创建时间"`
	UpdatedAt time.Time `json:"updatedAt" doc:"更新时间"`
}

type UserOutput struct {
	Body UserResponse
}

type TokenResponse struct {
	Token string `json:"token" doc:"JWT 访问令牌"`
}

type TokenOutput struct {
	Body TokenResponse
}

type SetupStatusResponse struct {
	Initialized bool `json:"initialized" doc:"是否已创建管理员"`
}

type SetupStatusOutput struct {
	Body SetupStatusResponse
}

type PasskeyStatusResponse struct {
	Available    bool   `json:"available" doc:"是否已有可用于登录的 Passkey"`
	Configured   bool   `json:"configured" doc:"Passkey 服务配置是否有效"`
	ErrorMessage string `json:"error_message,omitempty" doc:"配置错误说明"`
}

type PasskeyStatusOutput struct{ Body PasskeyStatusResponse }

type PasskeyRegistrationOptionsPayload struct {
	Name string `json:"name" minLength:"1" maxLength:"128" doc:"Passkey 名称"`
}

type PasskeyRegistrationOptionsInput struct {
	Body PasskeyRegistrationOptionsPayload
}

type PasskeyCreationOptionsResponse struct {
	SessionID string                      `json:"session_id"`
	Options   protocol.CredentialCreation `json:"options"`
}

type PasskeyCreationOptionsOutput struct {
	Body PasskeyCreationOptionsResponse
}

type PasskeyRequestOptionsResponse struct {
	SessionID string                       `json:"session_id"`
	Options   protocol.CredentialAssertion `json:"options"`
}

type PasskeyRequestOptionsOutput struct{ Body PasskeyRequestOptionsResponse }

type PasskeyVerifyPayload struct {
	SessionID  string          `json:"session_id" minLength:"1" maxLength:"128"`
	Credential json.RawMessage `json:"credential"`
}

type PasskeyVerifyInput struct{ Body PasskeyVerifyPayload }

type PasskeyResponse struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type PasskeyOutput struct{ Body PasskeyResponse }

type PasskeyListResponse struct {
	Items []PasskeyResponse `json:"items"`
}

type PasskeyListOutput struct{ Body PasskeyListResponse }

type PasskeyIDInput struct {
	ID string `path:"id"`
}

type PasskeyRenamePayload struct {
	Name string `json:"name" minLength:"1" maxLength:"128"`
}

type PasskeyRenameInput struct {
	ID   string `path:"id"`
	Body PasskeyRenamePayload
}

func newUserOutput(user model.User) *UserOutput {
	return &UserOutput{
		Body: UserResponse{
			ID:        user.Id,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}
}

func newPasskeyResponse(credential passkeymodel.Credential) PasskeyResponse {
	return PasskeyResponse{
		ID:         credential.Id,
		Name:       credential.Name,
		CreatedAt:  credential.CreatedAt,
		LastUsedAt: credential.LastUsedAt,
	}
}
