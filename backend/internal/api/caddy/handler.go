package caddy

import (
	"context"
	"strings"
	"time"

	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/db"
	"go-fiber-starter/pkg/logger"

	"github.com/danielgtaylor/huma/v2"
)

var newCaddyAdmin = func() service.CaddyAdmin { return service.NewCaddyClient() }

func Status(ctx context.Context, _ *struct{}) (*StatusOutput, error) {
	client := newCaddyAdmin()
	response := StatusResponse{AdminAPI: service.NewCaddyClient().AdminAPI}
	if concrete, ok := client.(*service.CaddyClient); ok {
		response.AdminAPI = concrete.AdminAPI
	}
	if err := client.GetStatus(ctx); err != nil {
		response.ErrorMessage = err.Error()
		return &StatusOutput{Body: response}, nil
	}
	response.Online = true
	return &StatusOutput{Body: response}, nil
}

func Version(ctx context.Context, _ *struct{}) (*VersionOutput, error) {
	info, err := service.NewCaddyVersionService().Check(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &VersionOutput{Body: VersionResponse{
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

func Update(ctx context.Context, input *UpdateInput) (*UpdateOutput, error) {
	manager := service.ManagedCaddy()
	if manager == nil {
		return nil, huma.Error503ServiceUnavailable("Caddy 托管服务尚未就绪")
	}
	target := strings.TrimSpace(input.Body.Version)
	if target == "" {
		info, err := service.NewCaddyVersionService().Check(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError(err.Error())
		}
		target = info.LatestVersion
	}
	if target == "" {
		return nil, huma.Error400BadRequest("无法确定 Caddy 目标版本")
	}

	go func(version string) {
		updateContext, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		if _, err := manager.Update(updateContext, version); err != nil {
			logger.Error("托管 Caddy 更新到 %s 失败: %v", version, err)
			return
		}
		logger.Info("托管 Caddy 已更新到 %s", version)
	}(target)

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
