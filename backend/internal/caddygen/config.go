package caddygen

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"

	"go-fiber-starter/internal/model/dnsprovider"
	"go-fiber-starter/internal/model/proxysite"
)

func Generate(sites []proxysite.ProxySite) ([]byte, error) {
	httpRoutes := make([]map[string]any, 0)
	httpsRoutes := make([]map[string]any, 0)
	loggedDomains := make(map[string]struct{})
	wildcardSubjects := make(map[string]struct{})
	tlsPolicies := newTLSPolicyAccumulator()
	for _, site := range sites {
		if !site.Enabled {
			continue
		}
		if site.EnableLog {
			domains, err := decodeStringList(site.Domains)
			if err != nil {
				return nil, fmt.Errorf("解析站点 %q 日志域名失败: %w", site.Name, err)
			}
			for _, domain := range domains {
				loggedDomains[domain] = struct{}{}
			}
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
				if err := tlsPolicies.Add(policy); err != nil {
					return nil, fmt.Errorf("生成站点 %q TLS 配置失败: %w", site.Name, err)
				}
				if site.CertificateType == "wildcard" {
					for _, subject := range policy["subjects"].([]string) {
						wildcardSubjects[subject] = struct{}{}
					}
				}
			}
		}
	}

	servers := map[string]any{ManagementServerName: managementServer()}
	if len(httpRoutes) > 0 {
		servers["proxy-http"] = siteServer(":80", httpRoutes, loggedDomains)
	}
	if len(httpsRoutes) > 0 {
		servers["proxy-https"] = siteServer(":443", httpsRoutes, loggedDomains)
	}
	apps := map[string]any{"http": map[string]any{"servers": servers}}
	if policies := tlsPolicies.Values(); len(policies) > 0 {
		tlsApp := map[string]any{"automation": map[string]any{"policies": policies}}
		if len(wildcardSubjects) > 0 {
			subjects := make([]string, 0, len(wildcardSubjects))
			for subject := range wildcardSubjects {
				subjects = append(subjects, subject)
			}
			sort.Strings(subjects)
			tlsApp["certificates"] = map[string]any{"automate": subjects}
		}
		apps["tls"] = tlsApp
	}
	config := map[string]any{
		"admin": localAdminConfig(),
		"apps":  apps,
	}
	if len(loggedDomains) > 0 {
		config["logging"] = accessLoggingConfig()
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
	location, err := httpsRedirectLocation()
	if err != nil {
		return nil, fmt.Errorf("生成站点 %q 跳转失败: %w", site.Name, err)
	}
	return map[string]any{
		"match": []map[string]any{{"host": domains}},
		"handle": []map[string]any{{
			"handler":     "static_response",
			"headers":     map[string][]string{"Location": {location}},
			"status_code": 308,
		}},
		"terminal": true,
	}, nil
}

func httpsRedirectLocation() (string, error) {
	port, err := strconv.Atoi(environmentValue("CADDYPILOT_HTTPS_PORT", "443"))
	if err != nil || port < 1 || port > 65535 {
		return "", fmt.Errorf("CADDYPILOT_HTTPS_PORT 必须是 1 到 65535 之间的端口")
	}
	location := "https://{http.request.host}"
	if port != 443 {
		location += ":" + strconv.Itoa(port)
	}
	return location + "{http.request.uri}", nil
}

func siteServer(listen string, routes []map[string]any, loggedDomains map[string]struct{}) map[string]any {
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
	if len(loggedDomains) > 0 {
		domains := make([]string, 0, len(loggedDomains))
		for domain := range loggedDomains {
			domains = append(domains, domain)
		}
		sort.Strings(domains)
		loggerNames := make(map[string][]string, len(domains))
		for _, domain := range domains {
			loggerNames[domain] = []string{"sites"}
		}
		server["logs"] = map[string]any{"logger_names": loggerNames}
	}
	return server
}

func accessLoggingConfig() map[string]any {
	return map[string]any{"logs": map[string]any{
		"default": map[string]any{"exclude": []string{"http.log.access.sites"}},
		"sites": map[string]any{
			"include": []string{"http.log.access.sites"},
			"encoder": map[string]any{"format": "json"},
			"writer": map[string]any{
				"output": "file", "filename": accessLogFilename(),
				"roll_size_mb": 20, "roll_keep": 5, "roll_keep_days": 30,
			},
		},
	}}
}

func accessLogFilename() string {
	return filepath.ToSlash(filepath.Join(environmentValue("CADDYPILOT_LOG_DIR", filepath.Join("data", "logs")), "access.log"))
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
