package caddy

import (
	"context"
	"os"
	"strings"

	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/db"
	"go-fiber-starter/pkg/logger"

	"github.com/danielgtaylor/huma/v2"
)

var newCaddyAdmin = func() service.CaddyAdmin {
	if manager := service.ManagedCaddy(); manager != nil {
		return manager
	}
	return service.NewCaddyClient()
}

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

	task, err := service.ManagedCaddyUpdateTasks().Start("download", target, func(updateContext context.Context, report func(string, int64, int64)) (string, error) {
		runtimeInfo, updateErr := manager.Update(updateContext, target, settings, report)
		if updateErr != nil {
			logger.Error("托管 Caddy 更新到 %s 失败: %v", target, updateErr)
			return "", updateErr
		}
		logger.Info("托管 Caddy 已更新到 %s", runtimeInfo.Version)
		return runtimeInfo.Version, nil
	})
	if err != nil {
		return nil, huma.Error409Conflict(err.Error())
	}

	return &UpdateOutput{Body: UpdateResponse{
		Accepted: true, TaskID: task.ID, Status: task.Status, TargetVersion: target,
	}}, nil
}

func Upload(_ context.Context, input *UploadInput) (*UpdateOutput, error) {
	manager := service.ManagedCaddy()
	if manager == nil {
		return nil, huma.Error503ServiceUnavailable("Caddy 托管服务尚未就绪")
	}
	file := input.RawBody.Data().File
	uploadPath, err := manager.Installer.SaveUpload(file)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	task, err := service.ManagedCaddyUpdateTasks().Start("upload", "", func(updateContext context.Context, report func(string, int64, int64)) (string, error) {
		runtimeInfo, updateErr := manager.UpdateUpload(updateContext, uploadPath, file.Filename, report)
		if updateErr != nil {
			logger.Error("安装上传的 Caddy 失败: %v", updateErr)
			return "", updateErr
		}
		logger.Info("已安装上传的 Caddy %s", runtimeInfo.Version)
		return runtimeInfo.Version, nil
	})
	if err != nil {
		_ = os.Remove(uploadPath)
		return nil, huma.Error409Conflict(err.Error())
	}
	return &UpdateOutput{Body: UpdateResponse{Accepted: true, TaskID: task.ID, Status: task.Status}}, nil
}

func CurrentUpdateTask(_ context.Context, _ *struct{}) (*UpdateTaskOutput, error) {
	task := service.ManagedCaddyUpdateTasks().Current()
	if task == nil {
		task = &service.CaddyUpdateTask{Status: "idle"}
	}
	return &UpdateTaskOutput{Body: *task}, nil
}

func Preview(ctx context.Context, _ *struct{}) (*JSONOutput, error) {
	payload, err := service.NewConfigService(db.DB, nil).Preview(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &JSONOutput{Body: JSONResponse{CaddyJSON: payload}}, nil
}

func ChangeStatus(ctx context.Context, _ *struct{}) (*ChangeStatusOutput, error) {
	status, err := service.NewConfigService(db.DB, newCaddyAdmin()).ChangeStatus(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &ChangeStatusOutput{Body: ChangeStatusResponse{
		Dirty: status.Dirty, State: status.State, LatestVersion: status.LatestVersion,
		LatestVersionID: status.LatestVersionID, ActiveVersion: status.ActiveVersion,
		RuntimeInSync: status.RuntimeInSync, PersistentConfigInSync: status.PersistentConfigInSync,
		ErrorMessage: status.ErrorMessage,
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
