package proxysite

import (
	"go-fiber-starter/internal/model/base"

	"gorm.io/gorm"
)

type ProxySite struct {
	base.BaseModel
	Name                          string         `gorm:"size:128;not null" json:"name"`
	Description                   string         `gorm:"type:text" json:"description"`
	Domains                       string         `gorm:"type:text;not null" json:"domains"`
	Upstreams                     string         `gorm:"type:text;not null" json:"upstreams"`
	UpstreamType                  string         `gorm:"size:16;not null;default:http" json:"upstream_type"`
	UpstreamTLSServerName         string         `gorm:"size:253" json:"upstream_tls_server_name"`
	UpstreamTLSInsecureSkipVerify bool           `gorm:"not null;default:false" json:"upstream_tls_insecure_skip_verify"`
	EnableHTTPS                   bool           `gorm:"column:enable_https;not null;default:false" json:"enable_https"`
	ForceHTTPS                    bool           `gorm:"column:force_https;not null;default:false" json:"force_https"`
	EnableGzip                    bool           `gorm:"not null;default:true" json:"enable_gzip"`
	EnableLog                     bool           `gorm:"not null;default:false" json:"enable_log"`
	EnableWS                      bool           `gorm:"not null;default:true" json:"enable_ws"`
	RequestHeaders                string         `gorm:"type:text;not null" json:"request_headers"`
	ResponseHeaders               string         `gorm:"type:text;not null" json:"response_headers"`
	BasicAuthEnabled              bool           `gorm:"not null;default:false" json:"basic_auth_enabled"`
	BasicAuthUsers                string         `gorm:"type:text;not null" json:"basic_auth_users"`
	BasicAuthCredentialIDs        string         `gorm:"type:text;not null;default:'[]'" json:"basic_auth_credential_ids"`
	AllowedIPs                    string         `gorm:"type:text;not null" json:"allowed_ips"`
	AdvancedJSON                  string         `gorm:"type:text" json:"advanced_json"`
	Enabled                       bool           `gorm:"not null;default:false;index" json:"enabled"`
	DeletedAt                     gorm.DeletedAt `gorm:"index" json:"-"`
}
