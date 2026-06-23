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
	RequestCount24h   int64      `json:"request_count_24h"`
	ErrorCount24h     int64      `json:"error_count_24h"`
	TrafficBytes24h   int64      `json:"traffic_bytes_24h"`
	TopSites24h       []TopSite  `json:"top_sites_24h"`
}

type TopSite struct {
	Domain       string `json:"domain"`
	RequestCount int64  `json:"request_count"`
	ErrorCount   int64  `json:"error_count"`
	Bytes        int64  `json:"bytes"`
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
	if managed := service.ManagedCaddy(); managed != nil {
		client = managed.Admin
	}
	response.CaddyOnline = client.GetStatus(ctx) == nil
	accessStats, err := service.LoadAccessStats(time.Now().Add(-24*time.Hour), 5)
	if err != nil {
		return nil, huma.Error500InternalServerError("读取访问统计失败", err)
	}
	response.RequestCount24h = accessStats.RequestCount
	response.ErrorCount24h = accessStats.ErrorCount
	response.TrafficBytes24h = accessStats.Bytes
	response.TopSites24h = make([]TopSite, 0, len(accessStats.TopSites))
	for _, site := range accessStats.TopSites {
		response.TopSites24h = append(response.TopSites24h, TopSite{
			Domain: site.Domain, RequestCount: site.RequestCount,
			ErrorCount: site.ErrorCount, Bytes: site.Bytes,
		})
	}
	return &SummaryOutput{Body: response}, nil
}
