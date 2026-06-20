package caddygen

import (
	"encoding/json"
	"fmt"

	"go-fiber-starter/internal/model/proxysite"
)

func Generate(sites []proxysite.ProxySite) ([]byte, error) {
	httpRoutes := make([]map[string]any, 0)
	httpsRoutes := make([]map[string]any, 0)
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
		}
	}

	servers := map[string]any{ManagementServerName: managementServer()}
	if len(httpRoutes) > 0 {
		servers["proxy-http"] = siteServer(":80", httpRoutes)
	}
	if len(httpsRoutes) > 0 {
		servers["proxy-https"] = siteServer(":443", httpsRoutes)
	}
	config := map[string]any{
		"admin": localAdminConfig(),
		"apps":  map[string]any{"http": map[string]any{"servers": servers}},
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
	return map[string]any{
		"listen": []string{listen},
		"routes": routes,
		"automatic_https": map[string]any{
			"disable_redirects": true,
		},
	}
}

func localAdminConfig() map[string]any {
	return map[string]any{
		"listen": "127.0.0.1:2019",
		"config": map[string]any{"persist": false},
	}
}
