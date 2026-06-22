package caddy

import (
	"encoding/json"

	"go-fiber-starter/internal/service"

	"github.com/danielgtaylor/huma/v2"
)

type JSONResponse struct {
	CaddyJSON json.RawMessage `json:"caddy_json"`
}

type JSONOutput struct {
	Body JSONResponse
}

type StatusResponse struct {
	Online       bool   `json:"online"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type StatusOutput struct {
	Body StatusResponse
}

type VersionResponse struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version,omitempty"`
	UpdateAvailable bool   `json:"update_available"`
	BinaryPath      string `json:"binary_path,omitempty"`
	VersionCheckURL string `json:"version_check_url"`
	DownloadURL     string `json:"download_url"`
	UpdateURL       string `json:"update_url,omitempty"`
	ReleaseURL      string `json:"release_url,omitempty"`
	UpdateStrategy  string `json:"update_strategy"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

type VersionOutput struct {
	Body VersionResponse
}

type SettingsPayload struct {
	VersionCheckURL string `json:"version_check_url" maxLength:"2048"`
	DownloadURL     string `json:"download_url" maxLength:"4096"`
	ChecksumURL     string `json:"checksum_url,omitempty" maxLength:"4096"`
}

type SettingsOutput struct{ Body SettingsPayload }
type SettingsInput struct{ Body SettingsPayload }

type UpdatePayload struct {
	Version string `json:"version,omitempty" doc:"目标 Caddy 版本，留空使用最新稳定版"`
}

type UpdateInput struct {
	Body UpdatePayload
}

type UpdateResponse struct {
	Accepted      bool   `json:"accepted"`
	TaskID        string `json:"task_id"`
	Status        string `json:"status"`
	TargetVersion string `json:"target_version"`
}

type UpdateOutput struct {
	Body UpdateResponse
}

type UploadForm struct {
	File huma.FormFile `form:"file" required:"true"`
}

type UploadInput struct {
	RawBody huma.MultipartFormFiles[UploadForm]
}

type UpdateTaskOutput struct {
	Body service.CaddyUpdateTask
}

type ValidateResponse struct {
	Valid bool `json:"valid"`
}

type ValidateOutput struct {
	Body ValidateResponse
}

type ChangeStatusResponse struct {
	Dirty         bool `json:"dirty"`
	LatestVersion uint `json:"latest_version,omitempty"`
}

type ChangeStatusOutput struct {
	Body ChangeStatusResponse
}

type PublishPayload struct {
	Reason string `json:"reason" maxLength:"255" doc:"发布原因"`
}

type PublishInput struct {
	Body PublishPayload
}

type PublishResponse struct {
	ID      uint   `json:"id"`
	Version uint   `json:"version"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
}

type PublishOutput struct {
	Body PublishResponse
}
