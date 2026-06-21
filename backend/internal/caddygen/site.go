package caddygen

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"go-fiber-starter/internal/model/proxysite"
)

func GenerateSiteRoute(site proxysite.ProxySite) (map[string]any, error) {
	domains, upstreamValues, allowedIPs, err := decodeSiteLists(site)
	if err != nil {
		return nil, err
	}
	requestHeaders, err := decodeStringMap(site.RequestHeaders)
	if err != nil {
		return nil, fmt.Errorf("解析请求头失败: %w", err)
	}
	responseHeaders, err := decodeStringMap(site.ResponseHeaders)
	if err != nil {
		return nil, fmt.Errorf("解析响应头失败: %w", err)
	}
	basicAuthUsers, err := decodeStringMap(site.BasicAuthUsers)
	if err != nil {
		return nil, fmt.Errorf("解析 Basic Auth 用户失败: %w", err)
	}

	matchers := map[string]any{"host": domains}
	if len(allowedIPs) > 0 {
		matchers["remote_ip"] = map[string]any{"ranges": allowedIPs}
	}

	handlers := make([]map[string]any, 0, 4)
	if len(responseHeaders) > 0 {
		handlers = append(handlers, responseHeaderHandler(responseHeaders))
	}
	if site.EnableGzip {
		handlers = append(handlers, encodeHandler())
	}
	if site.BasicAuthEnabled && len(basicAuthUsers) > 0 {
		handlers = append(handlers, basicAuthHandler(basicAuthUsers))
	}
	handlers = append(handlers, reverseProxyHandler(site, upstreamValues, requestHeaders))

	return map[string]any{
		"match": []map[string]any{matchers},
		"handle": []map[string]any{{
			"handler": "subroute",
			"routes":  []map[string]any{{"handle": handlers}},
		}},
		"terminal": true,
	}, nil
}

func decodeSiteLists(site proxysite.ProxySite) ([]string, []string, []string, error) {
	domains, err := decodeStringList(site.Domains)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("解析域名失败: %w", err)
	}
	if len(domains) == 0 {
		return nil, nil, nil, fmt.Errorf("域名不能为空")
	}
	upstreams, err := decodeStringList(site.Upstreams)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("解析上游失败: %w", err)
	}
	if len(upstreams) == 0 {
		return nil, nil, nil, fmt.Errorf("上游不能为空")
	}
	allowedIPs, err := decodeStringList(site.AllowedIPs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("解析 IP 白名单失败: %w", err)
	}
	return domains, upstreams, allowedIPs, nil
}

func reverseProxyHandler(site proxysite.ProxySite, upstreamValues []string, requestHeaders map[string]string) map[string]any {
	upstreams := make([]map[string]string, 0, len(upstreamValues))
	for _, upstream := range upstreamValues {
		upstreams = append(upstreams, map[string]string{"dial": upstreamDial(site.UpstreamType, upstream)})
	}
	handler := map[string]any{"handler": "reverse_proxy", "upstreams": upstreams}
	if transport := upstreamTransport(site); transport != nil {
		handler["transport"] = transport
	}
	if len(requestHeaders) > 0 {
		handler["headers"] = map[string]any{
			"request": map[string]any{"set": headerValues(requestHeaders)},
		}
	}
	return handler
}

func upstreamDial(upstreamType, value string) string {
	if upstreamType != "unix" || strings.HasPrefix(value, "unix/") {
		return value
	}
	return "unix/" + value
}

func upstreamTransport(site proxysite.ProxySite) map[string]any {
	switch site.UpstreamType {
	case "https":
		tlsConfig := map[string]any{}
		if site.UpstreamTLSServerName != "" {
			tlsConfig["server_name"] = site.UpstreamTLSServerName
		}
		if site.UpstreamTLSInsecureSkipVerify {
			tlsConfig["insecure_skip_verify"] = true
		}
		return map[string]any{"protocol": "http", "tls": tlsConfig}
	case "h2c":
		return map[string]any{"protocol": "http", "versions": []string{"h2c"}}
	default:
		return nil
	}
}

func responseHeaderHandler(headers map[string]string) map[string]any {
	return map[string]any{
		"handler":  "headers",
		"response": map[string]any{"set": headerValues(headers)},
	}
}

func encodeHandler() map[string]any {
	return map[string]any{
		"handler":   "encode",
		"encodings": map[string]any{"gzip": map[string]any{}, "zstd": map[string]any{}},
		"prefer":    []string{"gzip", "zstd"},
	}
}

func basicAuthHandler(users map[string]string) map[string]any {
	usernames := make([]string, 0, len(users))
	for username := range users {
		usernames = append(usernames, username)
	}
	sort.Strings(usernames)
	accounts := make([]map[string]string, 0, len(usernames))
	for _, username := range usernames {
		accounts = append(accounts, map[string]string{"username": username, "password": users[username]})
	}
	return map[string]any{
		"handler": "authentication",
		"providers": map[string]any{"http_basic": map[string]any{
			"accounts":   accounts,
			"hash":       map[string]string{"algorithm": "bcrypt"},
			"hash_cache": map[string]any{},
		}},
	}
}

func headerValues(headers map[string]string) map[string][]string {
	result := make(map[string][]string, len(headers))
	for name, value := range headers {
		result[name] = []string{value}
	}
	return result
}

func decodeStringList(value string) ([]string, error) {
	if value == "" {
		return []string{}, nil
	}
	var items []string
	if err := json.Unmarshal([]byte(value), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func decodeStringMap(value string) (map[string]string, error) {
	if value == "" {
		return map[string]string{}, nil
	}
	var items map[string]string
	if err := json.Unmarshal([]byte(value), &items); err != nil {
		return nil, err
	}
	return items, nil
}
