package proxysite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"go-fiber-starter/internal/caddygen"
	certificateModel "go-fiber-starter/internal/model/certificate"
	dnsProviderModel "go-fiber-starter/internal/model/dnsprovider"
	model "go-fiber-starter/internal/model/proxysite"
	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/db"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

var validateCaddyConfig = service.ValidateCaddyConfig

func List(_ context.Context, input *SiteListInput) (*SiteListOutput, error) {
	page, pageSize := input.Page, input.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var total int64
	if err := db.DB.Model(&model.ProxySite{}).Count(&total).Error; err != nil {
		return nil, huma.Error500InternalServerError("统计代理站点失败")
	}
	var sites []model.ProxySite
	if err := db.DB.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&sites).Error; err != nil {
		return nil, huma.Error500InternalServerError("查询代理站点失败")
	}

	items := make([]SiteResponse, 0, len(sites))
	for _, site := range sites {
		response, err := newSiteResponse(site)
		if err != nil {
			return nil, huma.Error500InternalServerError("解析代理站点失败")
		}
		items = append(items, response)
	}
	return &SiteListOutput{Body: SiteListResponse{
		Items: items, Total: total, Page: page, PageSize: pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}}, nil
}

func Create(ctx context.Context, input *SiteInput) (*SiteOutput, error) {
	site, err := siteFromPayload(input.Body)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	if err := validateCustomSite(ctx, site); err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	if err := db.DB.Create(&site).Error; err != nil {
		return nil, huma.Error500InternalServerError("创建代理站点失败", err)
	}
	return outputOrServerError(site)
}

func Get(_ context.Context, input *SiteIDInput) (*SiteOutput, error) {
	site, err := findSite(input.ID)
	if err != nil {
		return nil, err
	}
	return outputOrServerError(site)
}

func Update(ctx context.Context, input *UpdateSiteInput) (*SiteOutput, error) {
	site, err := findSite(input.ID)
	if err != nil {
		return nil, err
	}
	updated, err := siteFromPayload(input.Body)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	if err := validateCustomSite(ctx, updated); err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	updated.Id = site.Id
	updated.CreatedAt = site.CreatedAt
	if err := db.DB.Model(&site).Select("*").Omit("id", "created_at", "deleted_at").Updates(&updated).Error; err != nil {
		return nil, huma.Error500InternalServerError("更新代理站点失败", err)
	}
	return outputOrServerError(updated)
}

func Delete(_ context.Context, input *SiteIDInput) (*struct{}, error) {
	site, err := findSite(input.ID)
	if err != nil {
		return nil, err
	}
	if err := db.DB.Delete(&site).Error; err != nil {
		return nil, huma.Error500InternalServerError("删除代理站点失败", err)
	}
	return nil, nil
}

func Clone(ctx context.Context, input *CloneSiteInput) (*SiteOutput, error) {
	site, err := findSite(input.ID)
	if err != nil {
		return nil, err
	}
	site.Id = [16]byte{}
	site.CreatedAt = time.Time{}
	site.UpdatedAt = time.Time{}
	site.DeletedAt = gorm.DeletedAt{}
	site.Enabled = false
	if input.Body.Name != nil && strings.TrimSpace(*input.Body.Name) != "" {
		site.Name = strings.TrimSpace(*input.Body.Name)
	} else {
		site.Name += " 副本"
	}
	if len(input.Body.Domains) > 0 {
		site.Domains, _ = marshalJSON(input.Body.Domains)
	}
	if len(input.Body.Upstreams) > 0 {
		site.Upstreams, _ = marshalJSON(input.Body.Upstreams)
	}
	if site.Name == "" {
		return nil, huma.Error400BadRequest("站点名称不能为空")
	}
	if err := validateCustomSite(ctx, site); err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	if err := db.DB.Create(&site).Error; err != nil {
		return nil, huma.Error500InternalServerError("克隆代理站点失败")
	}
	return outputOrServerError(site)
}

func Enable(ctx context.Context, input *SiteIDInput) (*SiteOutput, error) {
	return setEnabled(ctx, input, true)
}

func Disable(ctx context.Context, input *SiteIDInput) (*SiteOutput, error) {
	return setEnabled(ctx, input, false)
}

func Preview(ctx context.Context, input *SiteIDInput) (*PreviewOutput, error) {
	site, err := findSite(input.ID)
	if err != nil {
		return nil, err
	}
	return previewSite(ctx, site)
}

func PreviewDraft(ctx context.Context, input *SiteInput) (*PreviewOutput, error) {
	site, err := siteFromPayload(input.Body)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	return previewSite(ctx, site)
}

func previewSite(ctx context.Context, site model.ProxySite) (*PreviewOutput, error) {
	sites := []model.ProxySite{site}
	if err := service.ResolveBasicAuthCredentials(ctx, db.DB, sites); err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	if err := service.ResolveCertificateProfiles(ctx, db.DB, sites); err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	route, err := caddygen.GenerateSiteRoute(sites[0])
	if err != nil {
		return nil, huma.Error500InternalServerError("生成站点配置失败")
	}
	payload, err := json.Marshal(route)
	if err != nil {
		return nil, huma.Error500InternalServerError("编码站点配置失败")
	}
	caddyfile, err := caddygen.GenerateSiteCaddyfile(sites[0])
	if err != nil {
		return nil, huma.Error500InternalServerError("生成站点 Caddyfile 失败")
	}
	return &PreviewOutput{Body: PreviewResponse{CaddyJSON: payload, Caddyfile: string(caddyfile)}}, nil
}

func validateCustomSite(ctx context.Context, site model.ProxySite) error {
	if site.ConfigMode != "custom" {
		return nil
	}
	site.Enabled = true
	payload, err := caddygen.Generate([]model.ProxySite{site})
	if err != nil {
		return fmt.Errorf("生成自定义站点校验配置失败: %w", err)
	}
	return validateCaddyConfig(ctx, payload)
}

func setEnabled(_ context.Context, input *SiteIDInput, enabled bool) (*SiteOutput, error) {
	site, err := findSite(input.ID)
	if err != nil {
		return nil, err
	}
	if err := db.DB.Model(&site).Update("enabled", enabled).Error; err != nil {
		return nil, huma.Error500InternalServerError("更新代理站点状态失败", err)
	}
	site.Enabled = enabled
	return outputOrServerError(site)
}

func findSite(id string) (model.ProxySite, error) {
	var site model.ProxySite
	if err := db.DB.First(&site, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return site, huma.Error404NotFound("代理站点不存在")
		}
		return site, huma.Error500InternalServerError("查询代理站点失败")
	}
	return site, nil
}

func siteFromPayload(payload SitePayload) (model.ProxySite, error) {
	configMode := strings.TrimSpace(payload.ConfigMode)
	if configMode == "" {
		configMode = "visual"
	}
	if configMode != "visual" && configMode != "custom" {
		return model.ProxySite{}, errors.New("配置模式无效")
	}
	customFormat := strings.TrimSpace(payload.CustomFormat)
	customConfig := strings.TrimSpace(payload.CustomConfig)
	customJSON := ""
	if configMode == "custom" {
		if customFormat != "json" && customFormat != "caddyfile" {
			return model.ProxySite{}, errors.New("自定义配置格式无效")
		}
		if customConfig == "" {
			return model.ProxySite{}, errors.New("自定义配置不能为空")
		}
		if customFormat == "json" {
			var route map[string]any
			if err := json.Unmarshal([]byte(customConfig), &route); err != nil {
				return model.ProxySite{}, errors.New("自定义 JSON 必须是有效的 Caddy 路由对象")
			}
			customJSON = customConfig
		} else {
			adapted, err := service.AdaptCaddyfile(context.Background(), []byte(customConfig))
			if err != nil {
				return model.ProxySite{}, err
			}
			route, err := firstAdaptedRoute(adapted)
			if err != nil {
				return model.ProxySite{}, err
			}
			routeJSON, _ := json.Marshal(route)
			customJSON = string(routeJSON)
		}
	}
	domains := compactStrings(payload.Domains)
	upstreams := compactStrings(payload.Upstreams)
	if configMode == "custom" {
		name := strings.TrimSpace(payload.Name)
		if name == "" {
			name = "自定义站点"
		}
		encodedDomains, _ := marshalJSON(domains)
		encodedUpstreams, _ := marshalJSON(upstreams)
		emptyObject, _ := marshalJSON(map[string]string{})
		emptyList, _ := marshalJSON([]string{})
		return model.ProxySite{
			Name: name, Description: strings.TrimSpace(payload.Description), SiteType: "proxy",
			ConfigMode: configMode, CustomFormat: customFormat, CustomConfig: customConfig, CustomJSON: customJSON,
			Domains: encodedDomains, Upstreams: encodedUpstreams, APIPath: "/api/*", UpstreamType: "http",
			EnableHTTPS: true, CertificateType: "single", ACMEChallengeType: "http", EnableGzip: true, EnableLog: true,
			RequestHeaders: emptyObject, ResponseHeaders: emptyObject, BasicAuthUsers: emptyObject,
			BasicAuthCredentialIDs: emptyList, AllowedIPs: emptyList, Enabled: payload.Enabled,
		}, nil
	}
	if len(domains) == 0 {
		return model.ProxySite{}, errors.New("域名不能为空")
	}
	siteType := normalizedSiteType(strings.TrimSpace(payload.SiteType))
	if !validSiteType(siteType) {
		return model.ProxySite{}, errors.New("不支持的站点类型")
	}
	rootPath := strings.TrimSpace(payload.RootPath)
	apiPath := defaultString(strings.TrimSpace(payload.APIPath), "/api/*")
	if siteType != "static" && len(upstreams) == 0 {
		return model.ProxySite{}, errors.New("反向代理和 SPA 站点必须配置上游")
	}
	if siteType != "proxy" && rootPath == "" {
		return model.ProxySite{}, errors.New("静态目录和 SPA 站点必须配置文件根目录")
	}
	if siteType == "spa" && !strings.HasPrefix(apiPath, "/") {
		return model.ProxySite{}, errors.New("API 路径必须以 / 开头")
	}
	upstreamType := normalizedUpstreamType(strings.TrimSpace(payload.UpstreamType))
	if !validUpstreamType(upstreamType) {
		return model.ProxySite{}, errors.New("不支持的上游类型")
	}
	certificateType := defaultString(strings.TrimSpace(payload.CertificateType), "single")
	challengeType := defaultString(strings.TrimSpace(payload.ACMEChallengeType), "http")
	certificateDomain := strings.TrimSpace(payload.CertificateDomain)
	if certificateType != "single" && certificateType != "wildcard" {
		return model.ProxySite{}, errors.New("不支持的证书类型")
	}
	if challengeType != "http" && challengeType != "dns" {
		return model.ProxySite{}, errors.New("不支持的 ACME 验证方式")
	}
	if certificateType == "wildcard" {
		challengeType = "dns"
		if payload.CertificateProfileID == nil {
			return model.ProxySite{}, errors.New("通配符证书必须选择系统证书配置")
		}
		var profile certificateModel.CertificateProfile
		if err := db.DB.Where("id = ? AND certificate_type = ? AND enabled = ?", *payload.CertificateProfileID, "wildcard", true).First(&profile).Error; err != nil {
			return model.ProxySite{}, errors.New("选择的通配符证书不存在或未启用")
		}
	}
	dnsProvider := ""
	if challengeType == "dns" {
		dnsProvider = defaultString(strings.TrimSpace(payload.DNSProvider), "alidns")
		if dnsProvider != "alidns" {
			return model.ProxySite{}, errors.New("当前只支持阿里云 DNS")
		}
		if certificateType == "single" {
			if payload.DNSProviderID == nil {
				return model.ProxySite{}, errors.New("DNS-01 必须选择系统 DNS Provider")
			}
			var provider dnsProviderModel.DNSProvider
			if err := db.DB.Where("id = ? AND enabled = ?", *payload.DNSProviderID, true).First(&provider).Error; err != nil {
				return model.ProxySite{}, errors.New("选择的 DNS Provider 不存在或未启用")
			}
		}
	}
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		name = domains[0]
	}

	encodedDomains, _ := marshalJSON(domains)
	encodedUpstreams, _ := marshalJSON(upstreams)
	requestHeaders, _ := marshalJSON(nonNilMap(payload.RequestHeaders))
	responseHeaders, _ := marshalJSON(nonNilMap(payload.ResponseHeaders))
	basicAuthUsers, _ := marshalJSON(map[string]string{})
	basicAuthCredentialIDs, _ := marshalJSON(payload.BasicAuthCredentialIDs)
	allowedIPs, _ := marshalJSON(compactStrings(payload.AllowedIPs))
	return model.ProxySite{
		Name:                          name,
		Description:                   strings.TrimSpace(payload.Description),
		SiteType:                      siteType,
		ConfigMode:                    configMode,
		CustomFormat:                  customFormat,
		CustomConfig:                  customConfig,
		CustomJSON:                    customJSON,
		Domains:                       encodedDomains,
		Upstreams:                     encodedUpstreams,
		RootPath:                      rootPath,
		APIPath:                       apiPath,
		EnableSecurityHeaders:         payload.EnableSecurityHeaders,
		EnableAssetCache:              payload.EnableAssetCache,
		UpstreamType:                  upstreamType,
		UpstreamTLSServerName:         strings.TrimSpace(payload.UpstreamTLSServerName),
		UpstreamTLSInsecureSkipVerify: payload.UpstreamTLSInsecureSkipVerify,
		EnableHTTPS:                   payload.EnableHTTPS,
		ForceHTTPS:                    payload.ForceHTTPS,
		CertificateType:               certificateType,
		CertificateDomain:             certificateDomain,
		ACMEChallengeType:             challengeType,
		DNSProvider:                   dnsProvider,
		DNSProviderID:                 payload.DNSProviderID,
		CertificateProfileID:          payload.CertificateProfileID,
		EnableGzip:                    payload.EnableGzip,
		EnableLog:                     payload.EnableLog,
		EnableWS:                      payload.EnableWS,
		RequestHeaders:                requestHeaders,
		ResponseHeaders:               responseHeaders,
		BasicAuthEnabled:              payload.BasicAuthEnabled,
		BasicAuthUsers:                basicAuthUsers,
		BasicAuthCredentialIDs:        basicAuthCredentialIDs,
		AllowedIPs:                    allowedIPs,
		AdvancedJSON:                  strings.TrimSpace(payload.AdvancedJSON),
		Enabled:                       payload.Enabled,
	}, nil
}

func firstAdaptedRoute(payload []byte) (map[string]any, error) {
	var config struct {
		Apps struct {
			HTTP struct {
				Servers map[string]struct {
					Routes []map[string]any `json:"routes"`
				} `json:"servers"`
			} `json:"http"`
		} `json:"apps"`
	}
	if err := json.Unmarshal(payload, &config); err != nil {
		return nil, errors.New("解析 Caddyfile 适配结果失败")
	}
	for _, server := range config.Apps.HTTP.Servers {
		if len(server.Routes) > 0 {
			return server.Routes[0], nil
		}
	}
	return nil, errors.New("Caddyfile 中没有可用的站点路由")
}

func validUpstreamType(value string) bool {
	switch value {
	case "http", "https", "h2c", "unix":
		return true
	default:
		return false
	}
}

func validSiteType(value string) bool {
	switch value {
	case "proxy", "static", "spa":
		return true
	default:
		return false
	}
}

func outputOrServerError(site model.ProxySite) (*SiteOutput, error) {
	output, err := newSiteOutput(site)
	if err != nil {
		return nil, huma.Error500InternalServerError("解析代理站点失败")
	}
	return output, nil
}

func marshalJSON(value any) (string, error) {
	payload, err := json.Marshal(value)
	return string(payload), err
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func nonNilMap(value map[string]string) map[string]string {
	if value == nil {
		return map[string]string{}
	}
	return value
}
