package proxysite

import (
	"encoding/json"
	"time"

	model "go-fiber-starter/internal/model/proxysite"

	"github.com/google/uuid"
)

type SitePayload struct {
	Name                          string            `json:"name,omitempty" maxLength:"128" doc:"兼容字段，留空时使用首个域名"`
	Description                   string            `json:"description" maxLength:"2000" doc:"站点描述"`
	SiteType                      string            `json:"site_type,omitempty" doc:"站点工作模式：proxy、static 或 spa"`
	ConfigMode                    string            `json:"config_mode,omitempty" doc:"配置模式：visual 或 custom"`
	CustomFormat                  string            `json:"custom_format,omitempty" doc:"自定义配置格式：json 或 caddyfile"`
	CustomConfig                  string            `json:"custom_config,omitempty" doc:"自定义站点配置"`
	Domains                       []string          `json:"domains" doc:"域名列表"`
	Upstreams                     []string          `json:"upstreams" doc:"上游地址列表"`
	RootPath                      string            `json:"root_path,omitempty" maxLength:"1024" doc:"静态文件根目录"`
	APIPath                       string            `json:"api_path,omitempty" maxLength:"256" doc:"SPA 的 API 路径匹配器"`
	EnableSecurityHeaders         bool              `json:"enable_security_headers,omitempty" doc:"启用常用安全响应头"`
	EnableAssetCache              bool              `json:"enable_asset_cache,omitempty" doc:"启用静态资源长期缓存"`
	UpstreamType                  string            `json:"upstream_type" enum:"http,https,h2c,unix" doc:"上游连接类型"`
	UpstreamTLSServerName         string            `json:"upstream_tls_server_name" doc:"HTTPS 上游 TLS Server Name"`
	UpstreamTLSInsecureSkipVerify bool              `json:"upstream_tls_insecure_skip_verify" doc:"跳过 HTTPS 上游证书校验"`
	EnableHTTPS                   bool              `json:"enable_https" doc:"启用 HTTPS"`
	ForceHTTPS                    bool              `json:"force_https" doc:"强制 HTTPS"`
	CertificateType               string            `json:"certificate_type" enum:"single,wildcard" doc:"证书类型"`
	CertificateDomain             string            `json:"certificate_domain" doc:"通配符证书域名"`
	ACMEChallengeType             string            `json:"acme_challenge_type" enum:"http,dns" doc:"ACME 验证方式"`
	DNSProvider                   string            `json:"dns_provider" enum:"alidns" doc:"DNS-01 服务商"`
	DNSProviderID                 *uuid.UUID        `json:"dns_provider_id,omitempty" doc:"系统 DNS Provider"`
	CertificateProfileID          *uuid.UUID        `json:"certificate_profile_id,omitempty" doc:"系统证书配置"`
	EnableGzip                    bool              `json:"enable_gzip" doc:"启用 gzip 和 zstd"`
	EnableLog                     bool              `json:"enable_log" doc:"启用访问日志"`
	EnableWS                      bool              `json:"enable_ws" doc:"启用 WebSocket"`
	RequestHeaders                map[string]string `json:"request_headers" doc:"请求头设置"`
	ResponseHeaders               map[string]string `json:"response_headers" doc:"响应头设置"`
	BasicAuthEnabled              bool              `json:"basic_auth_enabled" doc:"启用 Basic Auth"`
	BasicAuthUsers                map[string]string `json:"basic_auth_users" doc:"Basic Auth 用户及密码哈希"`
	BasicAuthCredentialIDs        []uuid.UUID       `json:"basic_auth_credential_ids" doc:"引用的密码本条目"`
	AllowedIPs                    []string          `json:"allowed_ips" doc:"允许访问的 IP 或网段"`
	AdvancedJSON                  string            `json:"advanced_json" doc:"暂存的高级 JSON"`
	Enabled                       bool              `json:"enabled" doc:"站点是否启用"`
}

type SiteInput struct {
	Body SitePayload
}

type SiteListInput struct {
	Page     int `query:"page" minimum:"1" default:"1"`
	PageSize int `query:"page_size" minimum:"1" maximum:"100" default:"20"`
}

type SiteIDInput struct {
	ID string `path:"id" format:"uuid" doc:"站点 ID"`
}

type UpdateSiteInput struct {
	ID   string `path:"id" format:"uuid" doc:"站点 ID"`
	Body SitePayload
}

type ClonePayload struct {
	Name      *string  `json:"name,omitempty" maxLength:"128" doc:"克隆站点名称"`
	Domains   []string `json:"domains,omitempty" doc:"覆盖域名列表"`
	Upstreams []string `json:"upstreams,omitempty" doc:"覆盖上游列表"`
}

type CloneSiteInput struct {
	ID   string `path:"id" format:"uuid" doc:"站点 ID"`
	Body ClonePayload
}

type SiteResponse struct {
	ID                            uuid.UUID         `json:"id"`
	Name                          string            `json:"name"`
	Description                   string            `json:"description"`
	SiteType                      string            `json:"site_type"`
	ConfigMode                    string            `json:"config_mode"`
	CustomFormat                  string            `json:"custom_format"`
	CustomConfig                  string            `json:"custom_config"`
	Domains                       []string          `json:"domains"`
	Upstreams                     []string          `json:"upstreams"`
	RootPath                      string            `json:"root_path"`
	APIPath                       string            `json:"api_path"`
	EnableSecurityHeaders         bool              `json:"enable_security_headers"`
	EnableAssetCache              bool              `json:"enable_asset_cache"`
	UpstreamType                  string            `json:"upstream_type"`
	UpstreamTLSServerName         string            `json:"upstream_tls_server_name"`
	UpstreamTLSInsecureSkipVerify bool              `json:"upstream_tls_insecure_skip_verify"`
	EnableHTTPS                   bool              `json:"enable_https"`
	ForceHTTPS                    bool              `json:"force_https"`
	CertificateType               string            `json:"certificate_type"`
	CertificateDomain             string            `json:"certificate_domain"`
	ACMEChallengeType             string            `json:"acme_challenge_type"`
	DNSProvider                   string            `json:"dns_provider"`
	DNSProviderID                 *uuid.UUID        `json:"dns_provider_id,omitempty"`
	CertificateProfileID          *uuid.UUID        `json:"certificate_profile_id,omitempty"`
	EnableGzip                    bool              `json:"enable_gzip"`
	EnableLog                     bool              `json:"enable_log"`
	EnableWS                      bool              `json:"enable_ws"`
	RequestHeaders                map[string]string `json:"request_headers"`
	ResponseHeaders               map[string]string `json:"response_headers"`
	BasicAuthEnabled              bool              `json:"basic_auth_enabled"`
	BasicAuthUsers                map[string]string `json:"basic_auth_users"`
	BasicAuthCredentialIDs        []uuid.UUID       `json:"basic_auth_credential_ids"`
	AllowedIPs                    []string          `json:"allowed_ips"`
	AdvancedJSON                  string            `json:"advanced_json"`
	Enabled                       bool              `json:"enabled"`
	CreatedAt                     time.Time         `json:"created_at"`
	UpdatedAt                     time.Time         `json:"updated_at"`
}

type SiteOutput struct {
	Body SiteResponse
}

type SiteListOutput struct {
	Body SiteListResponse
}

type SiteListResponse struct {
	Items      []SiteResponse `json:"items"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

type PreviewResponse struct {
	CaddyJSON json.RawMessage `json:"caddy_json"`
	Caddyfile string          `json:"caddyfile"`
}

type PreviewOutput struct {
	Body PreviewResponse
}

type NginxImportPayload struct {
	Config string `json:"config" doc:"Nginx 配置内容"`
}

type NginxImportInput struct {
	Body NginxImportPayload
}

type NginxImportResponse struct {
	Sites    []SiteResponse `json:"sites"`
	Warnings []string       `json:"warnings"`
}

type NginxImportOutput struct {
	Body NginxImportResponse
}

func newSiteOutput(site model.ProxySite) (*SiteOutput, error) {
	response, err := newSiteResponse(site)
	if err != nil {
		return nil, err
	}
	return &SiteOutput{Body: response}, nil
}

func newSiteResponse(site model.ProxySite) (SiteResponse, error) {
	response := SiteResponse{
		ID:                            site.Id,
		Name:                          site.Name,
		Description:                   site.Description,
		SiteType:                      normalizedSiteType(site.SiteType),
		ConfigMode:                    defaultString(site.ConfigMode, "visual"),
		CustomFormat:                  site.CustomFormat,
		CustomConfig:                  site.CustomConfig,
		RootPath:                      site.RootPath,
		APIPath:                       defaultString(site.APIPath, "/api/*"),
		EnableSecurityHeaders:         site.EnableSecurityHeaders,
		EnableAssetCache:              site.EnableAssetCache,
		EnableHTTPS:                   site.EnableHTTPS,
		UpstreamType:                  normalizedUpstreamType(site.UpstreamType),
		UpstreamTLSServerName:         site.UpstreamTLSServerName,
		UpstreamTLSInsecureSkipVerify: site.UpstreamTLSInsecureSkipVerify,
		ForceHTTPS:                    site.ForceHTTPS,
		CertificateType:               defaultString(site.CertificateType, "single"),
		CertificateDomain:             site.CertificateDomain,
		ACMEChallengeType:             defaultString(site.ACMEChallengeType, "http"),
		DNSProvider:                   defaultString(site.DNSProvider, "alidns"),
		DNSProviderID:                 site.DNSProviderID,
		CertificateProfileID:          site.CertificateProfileID,
		EnableGzip:                    site.EnableGzip,
		EnableLog:                     site.EnableLog,
		EnableWS:                      site.EnableWS,
		BasicAuthEnabled:              site.BasicAuthEnabled,
		BasicAuthUsers:                map[string]string{},
		AdvancedJSON:                  site.AdvancedJSON,
		Enabled:                       site.Enabled,
		CreatedAt:                     site.CreatedAt,
		UpdatedAt:                     site.UpdatedAt,
	}

	if err := decodeJSON(site.Domains, &response.Domains); err != nil {
		return SiteResponse{}, err
	}
	if err := decodeJSON(site.Upstreams, &response.Upstreams); err != nil {
		return SiteResponse{}, err
	}
	if err := decodeJSON(site.RequestHeaders, &response.RequestHeaders); err != nil {
		return SiteResponse{}, err
	}
	if err := decodeJSON(site.ResponseHeaders, &response.ResponseHeaders); err != nil {
		return SiteResponse{}, err
	}
	if err := decodeJSON(site.BasicAuthCredentialIDs, &response.BasicAuthCredentialIDs); err != nil {
		return SiteResponse{}, err
	}
	if err := decodeJSON(site.AllowedIPs, &response.AllowedIPs); err != nil {
		return SiteResponse{}, err
	}
	return response, nil
}

func decodeJSON(value string, target any) error {
	return json.Unmarshal([]byte(value), target)
}

func normalizedUpstreamType(value string) string {
	if value == "" {
		return "http"
	}
	return value
}

func normalizedSiteType(value string) string {
	return defaultString(value, "proxy")
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
