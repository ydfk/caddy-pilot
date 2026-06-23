package caddygen

import (
	"strings"
	"testing"

	"go-fiber-starter/internal/model/proxysite"
)

func TestGenerateCaddyfileIncludesManagementAndProxySites(t *testing.T) {
	t.Setenv("CADDYPILOT_HTTPS_PORT", "8443")
	site := proxysite.ProxySite{
		Name: "example.com", Domains: mustJSON([]string{"example.com"}),
		Upstreams: mustJSON([]string{"127.0.0.1:3000"}), UpstreamType: "http",
		RequestHeaders:  mustJSON(map[string]string{"X-Forwarded-Proto": "https"}),
		ResponseHeaders: mustJSON(map[string]string{}), BasicAuthUsers: mustJSON(map[string]string{}),
		AllowedIPs: mustJSON([]string{}), EnableHTTPS: true, ForceHTTPS: true, EnableGzip: true, Enabled: true,
	}
	payload, err := GenerateCaddyfile([]proxysite.ProxySite{site})
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, expected := range []string{":8080 {", "handle /api/*", "127.0.0.1:25610", "http://example.com {", "redir https://{host}:8443{uri} 308", "https://example.com {", "reverse_proxy 127.0.0.1:3000", "header_up X-Forwarded-Proto"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("Caddyfile 缺少 %q:\n%s", expected, text)
		}
	}
}

func TestGenerateCaddyfileMarksAdvancedJSONAsReadOnly(t *testing.T) {
	site := proxysite.ProxySite{
		Name: "advanced", Domains: mustJSON([]string{"advanced.example.com"}),
		Upstreams: mustJSON([]string{"127.0.0.1:3000"}), UpstreamType: "http",
		RequestHeaders: mustJSON(map[string]string{}), ResponseHeaders: mustJSON(map[string]string{}),
		BasicAuthUsers: mustJSON(map[string]string{}), AllowedIPs: mustJSON([]string{}),
		AdvancedJSON: `{"custom":true}`, Enabled: true,
	}
	payload, err := GenerateCaddyfile([]proxysite.ProxySite{site})
	if err != nil || !strings.Contains(string(payload), "仅能在 JSON 视图中完整表达") {
		t.Fatalf("高级 JSON 提示缺失: %v\n%s", err, payload)
	}
}

func TestGenerateCaddyfileIncludesStaticAndSPAContent(t *testing.T) {
	staticSite := testSite(true, nil)
	staticSite.Name = "静态站点"
	staticSite.Domains = mustJSON([]string{"static.example.com"})
	staticSite.SiteType = "static"
	staticSite.RootPath = "/var/www/example"
	spaSite := testSite(true, []string{"127.0.0.1:3000"})
	spaSite.Name = "SPA 站点"
	spaSite.Domains = mustJSON([]string{"app.example.com"})
	spaSite.SiteType = "spa"
	spaSite.RootPath = "/var/www/app/dist"
	spaSite.APIPath = "/api/*"
	spaSite.EnableSecurityHeaders = true
	spaSite.EnableAssetCache = true

	payload, err := GenerateCaddyfile([]proxysite.ProxySite{staticSite, spaSite})
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, expected := range []string{
		`root * "/var/www/example"`, "file_server", "handle /api/*", "reverse_proxy 127.0.0.1:3000",
		`root * "/var/www/app/dist"`, "try_files {path} /index.html", "X-Content-Type-Options", "/assets/* Cache-Control",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("Caddyfile 缺少 %q:\n%s", expected, text)
		}
	}
}
