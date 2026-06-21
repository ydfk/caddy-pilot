package basicauth

import (
	"time"

	model "go-fiber-starter/internal/model/basicauth"

	"github.com/google/uuid"
)

type CredentialPayload struct {
	Name     string `json:"name" minLength:"1" maxLength:"128" doc:"密码条目名称"`
	Username string `json:"username" minLength:"1" maxLength:"128" doc:"Basic Auth 用户名"`
	Password string `json:"password" minLength:"6" maxLength:"256" doc:"明文密码，仅用于生成哈希"`
}

type UpdateCredentialPayload struct {
	Name     string `json:"name" minLength:"1" maxLength:"128"`
	Username string `json:"username" minLength:"1" maxLength:"128"`
	Password string `json:"password,omitempty" maxLength:"256" doc:"留空时不修改密码"`
}

type CredentialInput struct{ Body CredentialPayload }
type CredentialIDInput struct {
	ID string `path:"id" format:"uuid"`
}
type UpdateCredentialInput struct {
	ID   string `path:"id" format:"uuid"`
	Body UpdateCredentialPayload
}

type CredentialResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CredentialOutput struct{ Body CredentialResponse }
type CredentialListOutput struct{ Body []CredentialResponse }

func newCredentialResponse(credential model.Credential) CredentialResponse {
	return CredentialResponse{
		ID: credential.Id, Name: credential.Name, Username: credential.Username,
		CreatedAt: credential.CreatedAt, UpdatedAt: credential.UpdatedAt,
	}
}
