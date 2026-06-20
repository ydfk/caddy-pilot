package configversion

import (
	"context"
	"errors"

	model "go-fiber-starter/internal/model/configversion"
	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/db"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

var newCaddyAdmin = func() service.CaddyAdmin { return service.NewCaddyClient() }

func List(_ context.Context, _ *struct{}) (*VersionListOutput, error) {
	var versions []model.ConfigVersion
	if err := db.DB.Order("version DESC").Find(&versions).Error; err != nil {
		return nil, huma.Error500InternalServerError("查询配置版本失败")
	}
	output := &VersionListOutput{Body: make([]VersionSummary, 0, len(versions))}
	for _, version := range versions {
		output.Body = append(output.Body, newVersionSummary(version))
	}
	return output, nil
}

func Get(_ context.Context, input *VersionIDInput) (*VersionOutput, error) {
	version, err := findVersion(input.ID)
	if err != nil {
		return nil, err
	}
	return &VersionOutput{Body: newVersionDetail(version)}, nil
}

func Rollback(ctx context.Context, input *VersionIDInput) (*VersionOutput, error) {
	version, err := service.NewConfigService(db.DB, newCaddyAdmin()).Rollback(ctx, input.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, huma.Error404NotFound("配置版本不存在")
		}
		return nil, huma.Error502BadGateway(err.Error())
	}
	return &VersionOutput{Body: newVersionDetail(version)}, nil
}

func findVersion(id uint) (model.ConfigVersion, error) {
	var version model.ConfigVersion
	if err := db.DB.First(&version, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return version, huma.Error404NotFound("配置版本不存在")
		}
		return version, huma.Error500InternalServerError("查询配置版本失败")
	}
	return version, nil
}
