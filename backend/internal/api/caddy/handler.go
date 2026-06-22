package caddy

import (
	"context"
	"strings"
	"time"

	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/db"
	"go-fiber-starter/pkg/logger"
	buildversion "go-fiber-starter/pkg/version"

	"github.com/danielgtaylor/huma/v2"
)

var newCaddyAdmin = func() service.CaddyAdmin { return service.NewCaddyClient() }

func Status(ctx context.Context, _ *struct{}) (*StatusOutput, error) {
	client := newCaddyAdmin()
	response := StatusResponse{}
	if err := client.GetStatus(ctx); err != nil {
		response.ErrorMessage = err.Error()
		return &StatusOutput{Body: response}, nil
	}
	response.Online = true
	return &StatusOutput{Body: response}, nil
}

func Version(ctx context.Context, _ *struct{}) (*VersionOutput, error) {
	settings, err := service.LoadCaddySettings(db.DB)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	versionService := service.NewCaddyVersionService()
	versionService.ReleaseAPI = settings.VersionCheckURL
	versionService.DownloadURL = settings.DownloadURL
	info, err := versionService.Check(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &VersionOutput{Body: VersionResponse{
		SystemVersion:   buildversion.Current,
		CurrentVersion:  info.CurrentVersion,
		LatestVersion:   info.LatestVersion,
		UpdateAvailable: info.UpdateAvailable,
		BinaryPath:      info.BinaryPath,
		VersionCheckURL: info.VersionCheckURL,
		DownloadURL:     info.DownloadURL,
		UpdateURL:       info.UpdateURL,
		ReleaseURL:      info.ReleaseURL,
		UpdateStrategy:  "managed",
		ErrorMessage:    info.ErrorMessage,
	}}, nil
}

func Settings(_ context.Context, _ *struct{}) (*SettingsOutput, error) {
	settings, err := service.LoadCaddySettings(db.DB)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &SettingsOutput{Body: SettingsPayload{
		VersionCheckURL: settings.VersionCheckURL,
		DownloadURL:     settings.DownloadURL,
		ChecksumURL:     settings.ChecksumURL,
	}}, nil
}

func SaveSettings(_ context.Context, input *SettingsInput) (*SettingsOutput, error) {
	settings := service.CaddySettings{
		VersionCheckURL: input.Body.VersionCheckURL,
		DownloadURL:     input.Body.DownloadURL,
		ChecksumURL:     input.Body.ChecksumURL,
	}
	if err := service.SaveCaddySettings(db.DB, settings); err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	settings, err := service.LoadCaddySettings(db.DB)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &SettingsOutput{Body: SettingsPayload{
		VersionCheckURL: settings.VersionCheckURL,
		DownloadURL:     settings.DownloadURL,
		ChecksumURL:     settings.ChecksumURL,
	}}, nil
}

func Update(ctx context.Context, input *UpdateInput) (*UpdateOutput, error) {
	manager := service.ManagedCaddy()
	if manager == nil {
		return nil, huma.Error503ServiceUnavailable("Caddy 托管服务尚未就绪")
	}
	target := strings.TrimSpace(input.Body.Version)
	settings, err := service.LoadCaddySettings(db.DB)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	if target == "" {
		versionService := service.NewCaddyVersionService()
		versionService.ReleaseAPI = settings.VersionCheckURL
		versionService.DownloadURL = settings.DownloadURL
		info, err := versionService.Check(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		target = info.LatestVersion
	}
	if target == "" {
		return nil, huma.Error400BadRequest("无法确定 Caddy 目标版本")
	}

	go func(version string, settings service.CaddySettings) {
		updateContext, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		if _, err := manager.Update(updateContext, version, settings); err != nil {
			logger.Error("托管 Caddy 更新到 %s 失败: %v", version, err)
			return
		}
		logger.Info("托管 Caddy 已更新到 %s", version)
	}(target, settings)

	return &UpdateOutput{Body: UpdateResponse{
		Accepted: true, TargetVersion: target,
	}}, nil
}

func Preview(ctx context.Context, _ *struct{}) (*JSONOutput, error) {
	payload, err := service.NewConfigService(db.DB, nil).Preview(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &JSONOutput{Body: JSONResponse{CaddyJSON: payload}}, nil
}

func ChangeStatus(ctx context.Context, _ *struct{}) (*ChangeStatusOutput, error) {
	status, err := service.NewConfigService(db.DB, nil).ChangeStatus(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &ChangeStatusOutput{Body: ChangeStatusResponse{
		Dirty: status.Dirty, LatestVersion: status.LatestVersion,
	}}, nil
}

func Validate(ctx context.Context, _ *struct{}) (*ValidateOutput, error) {
	payload, err := service.NewConfigService(db.DB, nil).Preview(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	if err := service.ValidateCaddyJSON(payload); err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	return &ValidateOutput{Body: ValidateResponse{Valid: true}}, nil
}

func Publish(ctx context.Context, input *PublishInput) (*PublishOutput, error) {
	version, err := service.NewConfigService(db.DB, newCaddyAdmin()).Publish(ctx, input.Body.Reason)
	if err != nil {
		return nil, huma.Error502BadGateway(err.Error())
	}
	return &PublishOutput{Body: PublishResponse{
		ID: version.ID, Version: version.Version, Status: version.Status, Reason: version.Reason,
	}}, nil
}

func CurrentConfig(ctx context.Context, _ *struct{}) (*JSONOutput, error) {
	payload, err := newCaddyAdmin().GetConfig(ctx)
	if err != nil {
		return nil, huma.Error502BadGateway(err.Error())
	}
	return &JSONOutput{Body: JSONResponse{CaddyJSON: payload}}, nil
}
