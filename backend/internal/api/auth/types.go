package auth

import (
	"time"

	model "go-fiber-starter/internal/model/user"

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
