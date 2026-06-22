package service

import (
	"context"
	"fmt"

	"go-fiber-starter/internal/model/certificate"
	"go-fiber-starter/internal/model/dnsprovider"
	"go-fiber-starter/internal/model/proxysite"

	"gorm.io/gorm"
)

func ResolveCertificateProfiles(ctx context.Context, database *gorm.DB, sites []proxysite.ProxySite) error {
	for index := range sites {
		site := &sites[index]
		if !site.EnableHTTPS || site.ACMEChallengeType != "dns" {
			continue
		}
		if site.CertificateType == "wildcard" && site.CertificateProfileID != nil {
			var profile certificate.CertificateProfile
			if err := database.WithContext(ctx).Where("id = ? AND enabled = ?", *site.CertificateProfileID, true).First(&profile).Error; err != nil {
				return fmt.Errorf("站点 %s 引用的证书配置不存在或未启用", site.Name)
			}
			site.ResolvedCertificateSubjects = profile.Subjects
			site.DNSProviderID = profile.DNSProviderID
		}
		if site.DNSProviderID == nil {
			return fmt.Errorf("站点 %s 的 DNS-01 未配置 DNS Provider", site.Name)
		}
		var provider dnsprovider.DNSProvider
		if err := database.WithContext(ctx).Where("id = ? AND enabled = ?", *site.DNSProviderID, true).First(&provider).Error; err != nil {
			return fmt.Errorf("站点 %s 引用的 DNS Provider 不存在或未启用", site.Name)
		}
		if provider.ProviderType != "alidns" {
			return fmt.Errorf("站点 %s 使用了暂不支持的 DNS Provider", site.Name)
		}
	}
	return nil
}
