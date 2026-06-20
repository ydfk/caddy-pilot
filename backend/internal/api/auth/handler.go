package auth

import (
	"context"
	"errors"

	model "go-fiber-starter/internal/model/user"
	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/db"

	"github.com/danielgtaylor/huma/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var generateFromPassword = bcrypt.GenerateFromPassword

func Register(_ context.Context, input *CredentialsInput) (*UserOutput, error) {
	var userCount int64
	if err := db.DB.Model(&model.User{}).Count(&userCount).Error; err != nil {
		return nil, huma.Error500InternalServerError("检查管理员状态失败")
	}
	if userCount > 0 {
		return nil, huma.Error409Conflict("管理员已初始化")
	}

	hash, err := generateFromPassword([]byte(input.Body.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, huma.Error500InternalServerError("密码加密失败")
	}

	user := model.User{
		Username: input.Body.Username,
		Password: string(hash),
	}
	if err := db.DB.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, huma.Error409Conflict("用户名已存在")
		}
		return nil, huma.Error500InternalServerError("创建用户失败")
	}

	return newUserOutput(user), nil
}

func Login(_ context.Context, input *CredentialsInput) (*TokenOutput, error) {
	var user model.User
	if err := db.DB.Where("username = ?", input.Body.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, huma.Error401Unauthorized("用户名或密码错误")
		}
		return nil, huma.Error500InternalServerError("查询用户失败")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Body.Password)) != nil {
		return nil, huma.Error401Unauthorized("用户名或密码错误")
	}

	token, err := service.GenerateJWT(&user)
	if err != nil {
		return nil, huma.Error500InternalServerError("Token 生成失败")
	}

	return &TokenOutput{Body: TokenResponse{Token: token}}, nil
}

func Profile(ctx context.Context, _ *struct{}) (*UserOutput, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("认证信息无效")
	}

	user, err := db.GetUserById(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, huma.Error404NotFound("用户不存在")
		}
		return nil, huma.Error500InternalServerError("查询用户失败")
	}

	return newUserOutput(user), nil
}
