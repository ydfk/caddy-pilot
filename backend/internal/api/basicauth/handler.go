package basicauth

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	model "go-fiber-starter/internal/model/basicauth"
	"go-fiber-starter/internal/model/proxysite"
	"go-fiber-starter/pkg/db"

	"github.com/danielgtaylor/huma/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func List(_ context.Context, _ *struct{}) (*CredentialListOutput, error) {
	var credentials []model.Credential
	if err := db.DB.Order("created_at DESC").Find(&credentials).Error; err != nil {
		return nil, huma.Error500InternalServerError("查询密码本失败")
	}
	output := &CredentialListOutput{Body: make([]CredentialResponse, 0, len(credentials))}
	for _, credential := range credentials {
		output.Body = append(output.Body, newCredentialResponse(credential))
	}
	return output, nil
}

func Create(_ context.Context, input *CredentialInput) (*CredentialOutput, error) {
	name, username := strings.TrimSpace(input.Body.Name), strings.TrimSpace(input.Body.Username)
	if name == "" || username == "" {
		return nil, huma.Error400BadRequest("名称和用户名不能为空")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Body.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, huma.Error500InternalServerError("生成密码哈希失败")
	}
	credential := model.Credential{Name: name, Username: username, PasswordHash: string(hash)}
	if err := db.DB.Create(&credential).Error; err != nil {
		return nil, huma.Error500InternalServerError("创建密码条目失败", err)
	}
	return &CredentialOutput{Body: newCredentialResponse(credential)}, nil
}

func Update(_ context.Context, input *UpdateCredentialInput) (*CredentialOutput, error) {
	credential, err := findCredential(input.ID)
	if err != nil {
		return nil, err
	}
	credential.Name = strings.TrimSpace(input.Body.Name)
	credential.Username = strings.TrimSpace(input.Body.Username)
	if credential.Name == "" || credential.Username == "" {
		return nil, huma.Error400BadRequest("名称和用户名不能为空")
	}
	if input.Body.Password != "" {
		if len(input.Body.Password) < 6 {
			return nil, huma.Error400BadRequest("密码至少需要 6 个字符")
		}
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(input.Body.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			return nil, huma.Error500InternalServerError("生成密码哈希失败")
		}
		credential.PasswordHash = string(hash)
	}
	if err := db.DB.Save(&credential).Error; err != nil {
		return nil, huma.Error500InternalServerError("更新密码条目失败", err)
	}
	return &CredentialOutput{Body: newCredentialResponse(credential)}, nil
}

func Delete(_ context.Context, input *CredentialIDInput) (*struct{}, error) {
	credential, err := findCredential(input.ID)
	if err != nil {
		return nil, err
	}
	inUse, err := credentialInUse(credential.Id.String())
	if err != nil {
		return nil, huma.Error500InternalServerError("检查密码条目引用失败")
	}
	if inUse {
		return nil, huma.Error400BadRequest("密码条目仍被代理站点引用，无法删除")
	}
	if err := db.DB.Delete(&credential).Error; err != nil {
		return nil, huma.Error500InternalServerError("删除密码条目失败", err)
	}
	return nil, nil
}

func credentialInUse(id string) (bool, error) {
	var sites []proxysite.ProxySite
	if err := db.DB.Select("basic_auth_credential_ids").Find(&sites).Error; err != nil {
		return false, err
	}
	for _, site := range sites {
		var ids []string
		if json.Unmarshal([]byte(site.BasicAuthCredentialIDs), &ids) != nil {
			continue
		}
		for _, candidate := range ids {
			if candidate == id {
				return true, nil
			}
		}
	}
	return false, nil
}

func findCredential(id string) (model.Credential, error) {
	var credential model.Credential
	if err := db.DB.First(&credential, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return credential, huma.Error404NotFound("密码条目不存在")
		}
		return credential, huma.Error500InternalServerError("查询密码条目失败")
	}
	return credential, nil
}
