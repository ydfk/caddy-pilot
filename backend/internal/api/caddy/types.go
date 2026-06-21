package caddy

import "encoding/json"

type JSONResponse struct {
	CaddyJSON json.RawMessage `json:"caddy_json"`
}

type JSONOutput struct {
	Body JSONResponse
}

type StatusResponse struct {
	Online       bool   `json:"online"`
	AdminAPI     string `json:"admin_api"`
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

type UpdatePayload struct {
	Version string `json:"version,omitempty" doc:"目标 Caddy 版本，留空使用最新稳定版"`
}

type UpdateInput struct {
	Body UpdatePayload
}

type UpdateResponse struct {
	Accepted      bool   `json:"accepted"`
	TargetVersion string `json:"target_version"`
}

type UpdateOutput struct {
	Body UpdateResponse
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
