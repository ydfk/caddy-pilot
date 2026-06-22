package certificate

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	model "go-fiber-starter/internal/model/certificate"
	"go-fiber-starter/internal/model/dnsprovider"
	"go-fiber-starter/internal/model/proxysite"
	"go-fiber-starter/pkg/db"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func List(_ context.Context, _ *struct{}) (*CertificateProfileListOutput, error) {
	var profiles []model.CertificateProfile
	if err := db.DB.Order("created_at DESC").Find(&profiles).Error; err != nil {
		return nil, huma.Error500InternalServerError("查询证书配置失败")
	}
	output := &CertificateProfileListOutput{Body: make([]CertificateProfileResponse, 0, len(profiles))}
	for _, profile := range profiles {
		subjects, err := decodeSubjects(profile.Subjects)
		if err != nil {
			return nil, huma.Error500InternalServerError("解析证书域名失败")
		}
		output.Body = append(output.Body, responseFromModel(profile, subjects))
	}
	return output, nil
}

func Create(_ context.Context, input *CertificateProfileCreateInput) (*CertificateProfileOutput, error) {
	profile, subjects, err := profileFromPayload(input.Body)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	if err := db.DB.Create(&profile).Error; err != nil {
		return nil, huma.Error500InternalServerError("创建证书配置失败", err)
	}
	return &CertificateProfileOutput{Body: responseFromModel(profile, subjects)}, nil
}

func Update(_ context.Context, input *CertificateProfileUpdateInput) (*CertificateProfileOutput, error) {
	existing, err := find(input.ID)
	if err != nil {
		return nil, err
	}
	profile, subjects, payloadErr := profileFromPayload(input.Body)
	if payloadErr != nil {
		return nil, huma.Error400BadRequest(payloadErr.Error())
	}
	profile.Id, profile.CreatedAt = existing.Id, existing.CreatedAt
	if err := db.DB.Model(&existing).Select("*").Omit("id", "created_at", "deleted_at").Updates(&profile).Error; err != nil {
		return nil, huma.Error500InternalServerError("更新证书配置失败", err)
	}
	return &CertificateProfileOutput{Body: responseFromModel(profile, subjects)}, nil
}

func Delete(_ context.Context, input *CertificateProfileIDInput) (*struct{}, error) {
	profile, err := find(input.ID)
	if err != nil {
		return nil, err
	}
	var count int64
	if err := db.DB.Model(&proxysite.ProxySite{}).Where("certificate_profile_id = ?", profile.Id).Count(&count).Error; err != nil {
		return nil, huma.Error500InternalServerError("检查证书引用失败")
	}
	if count > 0 {
		return nil, huma.Error400BadRequest("证书配置仍被代理站点引用，无法删除")
	}
	if err := db.DB.Delete(&profile).Error; err != nil {
		return nil, huma.Error500InternalServerError("删除证书配置失败", err)
	}
	return nil, nil
}

func profileFromPayload(payload CertificateProfilePayload) (model.CertificateProfile, []string, error) {
	subjects := compactStrings(payload.Subjects)
	if len(subjects) == 0 {
		return model.CertificateProfile{}, nil, errors.New("证书域名不能为空")
	}
	if payload.CertificateType != "single" && payload.CertificateType != "wildcard" {
		return model.CertificateProfile{}, nil, errors.New("不支持的证书类型")
	}
	challengeType := payload.ChallengeType
	if payload.CertificateType == "wildcard" {
		challengeType = "dns"
		for _, subject := range subjects {
			if !strings.HasPrefix(subject, "*.") {
				return model.CertificateProfile{}, nil, errors.New("通配符证书域名必须以 *. 开头")
			}
		}
	}
	if challengeType != "http" && challengeType != "dns" {
		return model.CertificateProfile{}, nil, errors.New("不支持的验证方式")
	}
	if challengeType == "dns" {
		if payload.DNSProviderID == nil {
			return model.CertificateProfile{}, nil, errors.New("DNS-01 必须选择 DNS Provider")
		}
		var provider dnsprovider.DNSProvider
		if err := db.DB.Where("id = ? AND enabled = ?", *payload.DNSProviderID, true).First(&provider).Error; err != nil {
			return model.CertificateProfile{}, nil, errors.New("选择的 DNS Provider 不存在或未启用")
		}
	}
	encoded, _ := json.Marshal(subjects)
	return model.CertificateProfile{
		Name: strings.TrimSpace(payload.Name), CertificateType: payload.CertificateType,
		Subjects: string(encoded), ChallengeType: challengeType,
		DNSProviderID: payload.DNSProviderID, Enabled: payload.Enabled,
	}, subjects, nil
}

func find(id string) (model.CertificateProfile, error) {
	var profile model.CertificateProfile
	if err := db.DB.First(&profile, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return profile, huma.Error404NotFound("证书配置不存在")
		}
		return profile, huma.Error500InternalServerError("查询证书配置失败")
	}
	return profile, nil
}

func decodeSubjects(value string) ([]string, error) {
	var subjects []string
	err := json.Unmarshal([]byte(value), &subjects)
	return subjects, err
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}
