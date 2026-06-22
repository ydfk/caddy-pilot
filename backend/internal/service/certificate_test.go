package service

import (
	"context"
	"testing"

	"go-fiber-starter/internal/model/certificate"
	"go-fiber-starter/internal/model/dnsprovider"
	"go-fiber-starter/internal/model/proxysite"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestResolveWildcardCertificateProfile(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := database.AutoMigrate(&dnsprovider.DNSProvider{}, &certificate.CertificateProfile{}); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}
	provider := dnsprovider.DNSProvider{Name: "阿里云", ProviderType: "alidns", EncryptedConfig: "encrypted", Enabled: true}
	if err := database.Create(&provider).Error; err != nil {
		t.Fatal(err)
	}
	profile := certificate.CertificateProfile{
		Name: "泛域名", CertificateType: "wildcard", Subjects: `["*.example.com"]`,
		ChallengeType: "dns", DNSProviderID: &provider.Id, Enabled: true,
	}
	if err := database.Create(&profile).Error; err != nil {
		t.Fatal(err)
	}
	sites := []proxysite.ProxySite{{
		Name: "example.com", EnableHTTPS: true, CertificateType: "wildcard",
		ACMEChallengeType: "dns", CertificateProfileID: &profile.Id,
	}}
	if err := ResolveCertificateProfiles(context.Background(), database, sites); err != nil {
		t.Fatalf("解析证书配置失败: %v", err)
	}
	if sites[0].DNSProviderID == nil || *sites[0].DNSProviderID != provider.Id {
		t.Fatal("站点未继承 DNS Provider")
	}
	if sites[0].ResolvedCertificateSubjects != profile.Subjects {
		t.Fatal("站点未继承证书域名")
	}
}
