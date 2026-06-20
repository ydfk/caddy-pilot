package proxysite

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"go-fiber-starter/internal/caddygen"
	model "go-fiber-starter/internal/model/proxysite"
	"go-fiber-starter/pkg/db"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func List(_ context.Context, _ *struct{}) (*SiteListOutput, error) {
	var sites []model.ProxySite
	if err := db.DB.Order("created_at DESC").Find(&sites).Error; err != nil {
		return nil, huma.Error500InternalServerError("查询代理站点失败")
	}

	output := &SiteListOutput{Body: make([]SiteResponse, 0, len(sites))}
	for _, site := range sites {
		response, err := newSiteResponse(site)
		if err != nil {
			return nil, huma.Error500InternalServerError("解析代理站点失败")
		}
		output.Body = append(output.Body, response)
	}
	return output, nil
}

func Create(_ context.Context, input *SiteInput) (*SiteOutput, error) {
	site, err := siteFromPayload(input.Body)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	if err := db.DB.Create(&site).Error; err != nil {
		return nil, huma.Error500InternalServerError("创建代理站点失败")
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

func Update(_ context.Context, input *UpdateSiteInput) (*SiteOutput, error) {
	site, err := findSite(input.ID)
	if err != nil {
		return nil, err
	}
	updated, err := siteFromPayload(input.Body)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	updated.Id = site.Id
	updated.CreatedAt = site.CreatedAt
	if err := db.DB.Model(&site).Select("*").Omit("id", "created_at", "deleted_at").Updates(&updated).Error; err != nil {
		return nil, huma.Error500InternalServerError("更新代理站点失败")
	}
	return outputOrServerError(updated)
}

func Delete(_ context.Context, input *SiteIDInput) (*struct{}, error) {
	site, err := findSite(input.ID)
	if err != nil {
		return nil, err
	}
	if err := db.DB.Delete(&site).Error; err != nil {
		return nil, huma.Error500InternalServerError("删除代理站点失败")
	}
	return nil, nil
}

func Clone(_ context.Context, input *CloneSiteInput) (*SiteOutput, error) {
	site, err := findSite(input.ID)
	if err != nil {
		return nil, err
	}
	site.Id = [16]byte{}
	site.CreatedAt = time.Time{}
	site.UpdatedAt = time.Time{}
	site.DeletedAt = gorm.DeletedAt{}
	site.Enabled = false
	if input.Body.Name != nil {
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

func Preview(_ context.Context, input *SiteIDInput) (*PreviewOutput, error) {
	site, err := findSite(input.ID)
	if err != nil {
		return nil, err
	}
	route, err := caddygen.GenerateSiteRoute(site)
	if err != nil {
		return nil, huma.Error500InternalServerError("生成站点配置失败")
	}
	payload, err := json.Marshal(route)
	if err != nil {
		return nil, huma.Error500InternalServerError("编码站点配置失败")
	}
	return &PreviewOutput{Body: PreviewResponse{CaddyJSON: payload}}, nil
}

func setEnabled(_ context.Context, input *SiteIDInput, enabled bool) (*SiteOutput, error) {
	site, err := findSite(input.ID)
	if err != nil {
		return nil, err
	}
	if err := db.DB.Model(&site).Update("enabled", enabled).Error; err != nil {
		return nil, huma.Error500InternalServerError("更新代理站点状态失败")
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
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return model.ProxySite{}, errors.New("站点名称不能为空")
	}
	domains := compactStrings(payload.Domains)
	upstreams := compactStrings(payload.Upstreams)
	if len(domains) == 0 || len(upstreams) == 0 {
		return model.ProxySite{}, errors.New("域名和上游不能为空")
	}

	encodedDomains, _ := marshalJSON(domains)
	encodedUpstreams, _ := marshalJSON(upstreams)
	requestHeaders, _ := marshalJSON(nonNilMap(payload.RequestHeaders))
	responseHeaders, _ := marshalJSON(nonNilMap(payload.ResponseHeaders))
	basicAuthUsers, _ := marshalJSON(nonNilMap(payload.BasicAuthUsers))
	allowedIPs, _ := marshalJSON(compactStrings(payload.AllowedIPs))
	return model.ProxySite{
		Name:             name,
		Description:      strings.TrimSpace(payload.Description),
		Domains:          encodedDomains,
		Upstreams:        encodedUpstreams,
		EnableHTTPS:      payload.EnableHTTPS,
		ForceHTTPS:       payload.ForceHTTPS,
		EnableGzip:       payload.EnableGzip,
		EnableLog:        payload.EnableLog,
		EnableWS:         payload.EnableWS,
		RequestHeaders:   requestHeaders,
		ResponseHeaders:  responseHeaders,
		BasicAuthEnabled: payload.BasicAuthEnabled,
		BasicAuthUsers:   basicAuthUsers,
		AllowedIPs:       allowedIPs,
		AdvancedJSON:     strings.TrimSpace(payload.AdvancedJSON),
		Enabled:          payload.Enabled,
	}, nil
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
