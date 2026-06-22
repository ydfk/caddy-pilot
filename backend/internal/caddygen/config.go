package caddygen

import (
	"encoding/json"
	"fmt"

	"go-fiber-starter/internal/model/dnsprovider"
	"go-fiber-starter/internal/model/proxysite"
)

func Generate(sites []proxysite.ProxySite) ([]byte, error) {
	httpRoutes := make([]map[string]any, 0)
	httpsRoutes := make([]map[string]any, 0)
	tlsPolicies := make([]map[string]any, 0)
	for _, site := range sites {
		if !site.Enabled {
			continue
		}
		route, err := GenerateSiteRoute(site)
		if err != nil {
			return nil, fmt.Errorf("生成站点 %q 配置失败: %w", site.Name, err)
		}

		if !site.EnableHTTPS || !site.ForceHTTPS {
			httpRoutes = append(httpRoutes, route)
		} else {
			redirect, err := generateRedirectRoute(site)
			if err != nil {
				return nil, err
			}
			httpRoutes = append(httpRoutes, redirect)
		}
		if site.EnableHTTPS {
			httpsRoutes = append(httpsRoutes, route)
			policy, err := siteTLSPolicy(site)
			if err != nil {
				return nil, fmt.Errorf("生成站点 %q TLS 配置失败: %w", site.Name, err)
			}
			if policy != nil {
				tlsPolicies = append(tlsPolicies, policy)
			}
		}
	}

	servers := map[string]any{ManagementServerName: managementServer()}
	if len(httpRoutes) > 0 {
		servers["proxy-http"] = siteServer(":80", httpRoutes)
	}
	if len(httpsRoutes) > 0 {
		servers["proxy-https"] = siteServer(":443", httpsRoutes)
	}
	apps := map[string]any{"http": map[string]any{"servers": servers}}
	if len(tlsPolicies) > 0 {
		apps["tls"] = map[string]any{"automation": map[string]any{"policies": tlsPolicies}}
	}
	config := map[string]any{
		"admin": localAdminConfig(),
		"apps":  apps,
	}
	return json.MarshalIndent(config, "", "  ")
}

func generateRedirectRoute(site proxysite.ProxySite) (map[string]any, error) {
	domains, err := decodeStringList(site.Domains)
	if err != nil {
		return nil, fmt.Errorf("生成站点 %q 跳转失败: %w", site.Name, err)
	}
	if len(domains) == 0 {
		return nil, fmt.Errorf("生成站点 %q 跳转失败: 域名不能为空", site.Name)
	}
	return map[string]any{
		"match": []map[string]any{{"host": domains}},
		"handle": []map[string]any{{
			"handler":     "static_response",
			"headers":     map[string][]string{"Location": {"https://{http.request.host}{http.request.uri}"}},
			"status_code": 308,
		}},
		"terminal": true,
	}, nil
}

func siteServer(listen string, routes []map[string]any) map[string]any {
	server := map[string]any{
		"listen": []string{listen},
		"routes": routes,
		"automatic_https": map[string]any{
			"disable_redirects": true,
		},
	}
	if listen == ":443" {
		server["tls_connection_policies"] = []map[string]any{{}}
	}
	return server
}

func siteTLSPolicy(site proxysite.ProxySite) (map[string]any, error) {
	if site.ACMEChallengeType != "dns" {
		return nil, nil
	}
	domains, err := decodeStringList(site.Domains)
	if err != nil {
		return nil, err
	}
	subjects := domains
	if site.CertificateType == "wildcard" {
		if site.ResolvedCertificateSubjects != "" {
			subjects, err = decodeStringList(site.ResolvedCertificateSubjects)
			if err != nil {
				return nil, fmt.Errorf("解析证书配置域名失败: %w", err)
			}
		} else if site.CertificateDomain != "" {
			subjects = []string{site.CertificateDomain}
		} else {
			return nil, fmt.Errorf("通配符证书域名不能为空")
		}
	}
	provider := map[string]any{
		"name": "alidns", "access_key_id": "{env.ALIYUN_ACCESS_KEY_ID}",
		"access_key_secret": "{env.ALIYUN_ACCESS_KEY_SECRET}",
	}
	if site.DNSProviderID != nil {
		idName, secretName, regionName := dnsprovider.EnvNames(*site.DNSProviderID)
		provider["access_key_id"] = "{env." + idName + "}"
		provider["access_key_secret"] = "{env." + secretName + "}"
		provider["region_id"] = "{env." + regionName + "}"
	}
	issuer := map[string]any{
		"module":     "acme",
		"challenges": map[string]any{"dns": map[string]any{"provider": provider}},
	}
	return map[string]any{"subjects": subjects, "issuers": []map[string]any{issuer}}, nil
}

func localAdminConfig() map[string]any {
	return map[string]any{
		"listen": "127.0.0.1:2019",
		"config": map[string]any{"persist": false},
	}
}
