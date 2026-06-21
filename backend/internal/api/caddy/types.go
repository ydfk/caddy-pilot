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
	ReleaseURL      string `json:"release_url,omitempty"`
	UpdateCommand   string `json:"update_command,omitempty"`
	UpdateStrategy  string `json:"update_strategy"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

type VersionOutput struct {
	Body VersionResponse
}

type ValidateResponse struct {
	Valid bool `json:"valid"`
}

type ValidateOutput struct {
	Body ValidateResponse
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
