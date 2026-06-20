package configversion

import (
	"encoding/json"
	"time"

	model "go-fiber-starter/internal/model/configversion"
)

type VersionIDInput struct {
	ID uint `path:"id" minimum:"1" doc:"配置版本 ID"`
}

type VersionSummary struct {
	ID          uint       `json:"id"`
	Version     uint       `json:"version"`
	Reason      string     `json:"reason"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type VersionDetail struct {
	VersionSummary
	BusinessConfig json.RawMessage `json:"business_config"`
	CaddyJSON      json.RawMessage `json:"caddy_json"`
	ErrorMessage   string          `json:"error_message"`
}

type VersionListOutput struct {
	Body []VersionSummary
}

type VersionOutput struct {
	Body VersionDetail
}

func newVersionSummary(version model.ConfigVersion) VersionSummary {
	return VersionSummary{
		ID: version.ID, Version: version.Version, Reason: version.Reason,
		Status: version.Status, PublishedAt: version.PublishedAt, CreatedAt: version.CreatedAt,
	}
}

func newVersionDetail(version model.ConfigVersion) VersionDetail {
	return VersionDetail{
		VersionSummary: newVersionSummary(version),
		BusinessConfig: safeRawJSON(version.BusinessConfig),
		CaddyJSON:      safeRawJSON(version.CaddyJSON),
		ErrorMessage:   version.ErrorMessage,
	}
}

func safeRawJSON(value string) json.RawMessage {
	payload := []byte(value)
	if !json.Valid(payload) {
		return json.RawMessage("null")
	}
	return payload
}
