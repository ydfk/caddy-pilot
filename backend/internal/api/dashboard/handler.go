package dashboard

import (
	"context"
	"errors"
	"time"

	"go-fiber-starter/internal/model/configversion"
	"go-fiber-starter/internal/model/proxysite"
	"go-fiber-starter/internal/service"
	"go-fiber-starter/pkg/db"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

type SummaryResponse struct {
	SiteCount         int64      `json:"site_count"`
	EnabledSiteCount  int64      `json:"enabled_site_count"`
	DisabledSiteCount int64      `json:"disabled_site_count"`
	HTTPSSiteCount    int64      `json:"https_site_count"`
	LastPublishTime   *time.Time `json:"last_publish_time"`
	CaddyOnline       bool       `json:"caddy_online"`
}

type SummaryOutput struct {
	Body SummaryResponse
}

func Summary(ctx context.Context, _ *struct{}) (*SummaryOutput, error) {
	response := SummaryResponse{}
	if err := db.DB.WithContext(ctx).Model(&proxysite.ProxySite{}).Count(&response.SiteCount).Error; err != nil {
		return nil, huma.Error500InternalServerError("统计代理站点失败")
	}
	if err := db.DB.WithContext(ctx).Model(&proxysite.ProxySite{}).Where("enabled = ?", true).Count(&response.EnabledSiteCount).Error; err != nil {
		return nil, huma.Error500InternalServerError("统计启用站点失败")
	}
	response.DisabledSiteCount = response.SiteCount - response.EnabledSiteCount
	if err := db.DB.WithContext(ctx).Model(&proxysite.ProxySite{}).Where("enable_https = ?", true).Count(&response.HTTPSSiteCount).Error; err != nil {
		return nil, huma.Error500InternalServerError("统计 HTTPS 站点失败")
	}

	var latest configversion.ConfigVersion
	result := db.DB.WithContext(ctx).Where("status IN ?", []string{service.ConfigStatusPublished, service.ConfigStatusRollback}).Order("version DESC").First(&latest)
	if result.Error == nil {
		response.LastPublishTime = latest.PublishedAt
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, huma.Error500InternalServerError("查询最近发布时间失败")
	}
	client := service.NewCaddyClient()
	response.CaddyOnline = client.GetStatus(ctx) == nil
	return &SummaryOutput{Body: response}, nil
}
