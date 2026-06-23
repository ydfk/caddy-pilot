package proxysite

import (
	"context"
	"fmt"
	"strings"

	model "go-fiber-starter/internal/model/proxysite"
	"go-fiber-starter/internal/nginximport"
	"go-fiber-starter/pkg/db"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

const maxNginxImportSize = 2 << 20

func ImportNginx(ctx context.Context, input *NginxImportInput) (*NginxImportOutput, error) {
	config := strings.TrimSpace(input.Body.Config)
	if config == "" {
		return nil, huma.Error400BadRequest("Nginx 配置不能为空")
	}
	if len(config) > maxNginxImportSize {
		return nil, huma.Error400BadRequest("Nginx 配置不能超过 2 MiB")
	}
	parsed, err := nginximport.Parse(config)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}

	models := make([]model.ProxySite, 0, len(parsed.Sites))
	for _, imported := range parsed.Sites {
		site, convertErr := siteFromPayload(importedSitePayload(imported))
		if convertErr != nil {
			return nil, huma.Error400BadRequest(fmt.Sprintf("转换站点 %s 失败: %v", strings.Join(imported.Domains, ", "), convertErr))
		}
		models = append(models, site)
	}
	if err := db.DB.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		return transaction.Create(&models).Error
	}); err != nil {
		return nil, huma.Error500InternalServerError("导入 Nginx 站点失败", err)
	}

	responses := make([]SiteResponse, 0, len(models))
	for _, site := range models {
		response, responseErr := newSiteResponse(site)
		if responseErr != nil {
			return nil, huma.Error500InternalServerError("解析已导入站点失败", responseErr)
		}
		responses = append(responses, response)
	}
	return &NginxImportOutput{Body: NginxImportResponse{Sites: responses, Warnings: parsed.Warnings}}, nil
}

func importedSitePayload(site nginximport.Site) SitePayload {
	return SitePayload{
		Name: site.Domains[0], Description: "从 Nginx 配置导入",
		SiteType: "proxy",
		Domains:  site.Domains, Upstreams: site.Upstreams, UpstreamType: site.UpstreamType,
		EnableHTTPS: site.EnableHTTPS, ForceHTTPS: site.ForceHTTPS,
		CertificateType: "single", ACMEChallengeType: "http",
		EnableGzip: site.EnableGzip, EnableLog: site.EnableLog, EnableWS: true,
		RequestHeaders: map[string]string{}, ResponseHeaders: map[string]string{},
		BasicAuthUsers: map[string]string{}, BasicAuthCredentialIDs: nil, AllowedIPs: nil,
		Enabled: false,
	}
}
