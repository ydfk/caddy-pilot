package caddygen

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"go-fiber-starter/internal/model/dnsprovider"
	"go-fiber-starter/internal/model/proxysite"
)

func GenerateCaddyfile(sites []proxysite.ProxySite) ([]byte, error) {
	var output strings.Builder
	writeGlobalOptions(&output)
	writeManagementSite(&output)
	for _, site := range sites {
		if !site.Enabled {
			continue
		}
		if err := writeProxySite(&output, site); err != nil {
			return nil, fmt.Errorf("生成站点 %q Caddyfile 失败: %w", site.Name, err)
		}
	}
	return []byte(output.String()), nil
}

func GenerateSiteCaddyfile(site proxysite.ProxySite) ([]byte, error) {
	var output strings.Builder
	site.Enabled = true
	if err := writeProxySite(&output, site); err != nil {
		return nil, err
	}
	return []byte(output.String()), nil
}

func writeGlobalOptions(output *strings.Builder) {
	output.WriteString("{\n")
	output.WriteString("\tadmin 127.0.0.1:2019\n")
	output.WriteString("\tpersist_config off\n")
	output.WriteString("}\n\n")
}

func writeManagementSite(output *strings.Builder) {
	fmt.Fprintf(output, "%s {\n", managementListen())
	if frontendProxy := normalizeDial(environmentValue("CADDYPILOT_FRONTEND_PROXY", "")); frontendProxy != "" {
		fmt.Fprintf(output, "\treverse_proxy %s\n", frontendProxy)
		output.WriteString("}\n\n")
		return
	}
	fmt.Fprintf(output, "\troot * %s\n", quoteCaddyfile(frontendRoot()))
	output.WriteString("\tencode gzip zstd\n")
	output.WriteString("\thandle /api/* {\n")
	fmt.Fprintf(output, "\t\treverse_proxy %s\n", backendUpstream())
	output.WriteString("\t}\n")
	output.WriteString("\thandle {\n")
	output.WriteString("\t\ttry_files {path} /index.html\n")
	output.WriteString("\t\tfile_server\n")
	output.WriteString("\t}\n")
	output.WriteString("}\n\n")
}

func writeProxySite(output *strings.Builder, site proxysite.ProxySite) error {
	domains, upstreams, allowedIPs, err := decodeSiteLists(site)
	if err != nil {
		return err
	}
	addresses := siteAddresses(domains, site.EnableHTTPS, site.ForceHTTPS)
	fmt.Fprintf(output, "%s {\n", strings.Join(addresses, ", "))
	if site.EnableGzip {
		output.WriteString("\tencode gzip zstd\n")
	}
	if len(allowedIPs) > 0 {
		fmt.Fprintf(output, "\t@allowed remote_ip %s\n", strings.Join(allowedIPs, " "))
		output.WriteString("\thandle @allowed {\n")
		if err := writeSiteHandlers(output, site, upstreams, "\t\t"); err != nil {
			return err
		}
		output.WriteString("\t}\n")
		output.WriteString("\trespond 403\n")
	} else if err := writeSiteHandlers(output, site, upstreams, "\t"); err != nil {
		return err
	}
	if site.EnableHTTPS && site.ACMEChallengeType == "dns" {
		writeDNSIssuer(output, site)
	}
	if strings.TrimSpace(site.AdvancedJSON) != "" {
		output.WriteString("\t# advanced_json 仅能在 JSON 视图中完整表达\n")
	}
	output.WriteString("}\n\n")
	return nil
}

func writeSiteHandlers(output *strings.Builder, site proxysite.ProxySite, upstreams []string, indent string) error {
	responseHeaders, err := decodeStringMap(site.ResponseHeaders)
	if err != nil {
		return fmt.Errorf("解析响应头失败: %w", err)
	}
	for _, key := range sortedKeys(responseHeaders) {
		fmt.Fprintf(output, "%sheader %s %s\n", indent, key, quoteCaddyfile(responseHeaders[key]))
	}
	if site.BasicAuthEnabled {
		users, err := decodeStringMap(site.BasicAuthUsers)
		if err != nil {
			return fmt.Errorf("解析 Basic Auth 用户失败: %w", err)
		}
		if len(users) > 0 {
			fmt.Fprintf(output, "%sbasic_auth {\n", indent)
			for _, username := range sortedKeys(users) {
				fmt.Fprintf(output, "%s\t%s %s\n", indent, quoteCaddyfile(username), users[username])
			}
			fmt.Fprintf(output, "%s}\n", indent)
		}
	}
	fmt.Fprintf(output, "%sreverse_proxy %s", indent, strings.Join(caddyfileUpstreams(site.UpstreamType, upstreams), " "))
	requestHeaders, err := decodeStringMap(site.RequestHeaders)
	if err != nil {
		return fmt.Errorf("解析请求头失败: %w", err)
	}
	needsBlock := len(requestHeaders) > 0 || site.UpstreamType == "https"
	if !needsBlock {
		output.WriteString("\n")
		return nil
	}
	output.WriteString(" {\n")
	for _, key := range sortedKeys(requestHeaders) {
		fmt.Fprintf(output, "%s\theader_up %s %s\n", indent, key, quoteCaddyfile(requestHeaders[key]))
	}
	if site.UpstreamType == "https" {
		fmt.Fprintf(output, "%s\ttransport http {\n", indent)
		output.WriteString(indent + "\t\ttls\n")
		if site.UpstreamTLSServerName != "" {
			fmt.Fprintf(output, "%s\t\ttls_server_name %s\n", indent, quoteCaddyfile(site.UpstreamTLSServerName))
		}
		if site.UpstreamTLSInsecureSkipVerify {
			output.WriteString(indent + "\t\ttls_insecure_skip_verify\n")
		}
		fmt.Fprintf(output, "%s\t}\n", indent)
	}
	fmt.Fprintf(output, "%s}\n", indent)
	return nil
}

func writeDNSIssuer(output *strings.Builder, site proxysite.ProxySite) {
	idName, secretName, regionName := "ALIYUN_ACCESS_KEY_ID", "ALIYUN_ACCESS_KEY_SECRET", ""
	if site.DNSProviderID != nil {
		idName, secretName, regionName = dnsprovider.EnvNames(*site.DNSProviderID)
	}
	output.WriteString("\ttls {\n")
	output.WriteString("\t\tdns alidns {\n")
	fmt.Fprintf(output, "\t\t\taccess_key_id {env.%s}\n", idName)
	fmt.Fprintf(output, "\t\t\taccess_key_secret {env.%s}\n", secretName)
	if regionName != "" {
		fmt.Fprintf(output, "\t\t\tregion_id {env.%s}\n", regionName)
	}
	output.WriteString("\t\t}\n")
	output.WriteString("\t}\n")
	if site.CertificateType == "wildcard" {
		output.WriteString("\t# 通配符证书的 subjects 策略以 JSON 视图为准\n")
	}
}

func siteAddresses(domains []string, https, forceHTTPS bool) []string {
	addresses := make([]string, 0, len(domains)*2)
	for _, domain := range domains {
		switch {
		case !https:
			addresses = append(addresses, "http://"+domain)
		case forceHTTPS:
			addresses = append(addresses, domain)
		default:
			addresses = append(addresses, "http://"+domain, "https://"+domain)
		}
	}
	return addresses
}

func caddyfileUpstreams(kind string, upstreams []string) []string {
	result := make([]string, 0, len(upstreams))
	for _, upstream := range upstreams {
		switch kind {
		case "https":
			result = append(result, "https://"+strings.TrimPrefix(upstream, "https://"))
		case "h2c":
			result = append(result, "h2c://"+strings.TrimPrefix(upstream, "h2c://"))
		case "unix":
			result = append(result, "unix/"+strings.TrimPrefix(upstream, "unix/"))
		default:
			result = append(result, upstream)
		}
	}
	return result
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func quoteCaddyfile(value string) string {
	return strconv.Quote(value)
}
