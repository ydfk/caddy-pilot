package caddygen

import (
	"bytes"
	"encoding/json"
	"testing"

	"go-fiber-starter/internal/model/dnsprovider"
	"go-fiber-starter/internal/model/proxysite"

	"github.com/google/uuid"
)

func TestGenerateEmptySitesContainsManagementEntry(t *testing.T) {
	payload, err := Generate(nil)
	if err != nil {
		t.Fatalf("生成空站点配置失败: %v", err)
	}
	if !HasManagementEntry(payload) {
		t.Fatalf("空站点配置缺少管理入口: %s", payload)
	}
}

func TestGenerateSiteContainsHostAndReverseProxy(t *testing.T) {
	payload, err := Generate([]proxysite.ProxySite{testSite(true, []string{"127.0.0.1:3000"})})
	if err != nil {
		t.Fatalf("生成站点配置失败: %v", err)
	}
	for _, expected := range [][]byte{[]byte("example.com"), []byte("reverse_proxy"), []byte("127.0.0.1:3000")} {
		if !bytes.Contains(payload, expected) {
			t.Fatalf("生成配置缺少 %s: %s", expected, payload)
		}
	}
}

func TestGenerateSkipsDisabledSite(t *testing.T) {
	payload, err := Generate([]proxysite.ProxySite{testSite(false, []string{"127.0.0.1:3000"})})
	if err != nil {
		t.Fatalf("生成站点配置失败: %v", err)
	}
	if bytes.Contains(payload, []byte("example.com")) {
		t.Fatalf("停用站点进入了生成配置: %s", payload)
	}
}

func TestGenerateSupportsMultipleUpstreams(t *testing.T) {
	payload, err := Generate([]proxysite.ProxySite{testSite(true, []string{"127.0.0.1:3000", "127.0.0.1:3001"})})
	if err != nil {
		t.Fatalf("生成多上游配置失败: %v", err)
	}
	for _, upstream := range [][]byte{[]byte("127.0.0.1:3000"), []byte("127.0.0.1:3001")} {
		if !bytes.Contains(payload, upstream) {
			t.Fatalf("生成配置缺少上游 %s: %s", upstream, payload)
		}
	}
}

func TestGenerateSupportsStaticFileSite(t *testing.T) {
	site := testSite(true, nil)
	site.SiteType = "static"
	site.RootPath = "/var/www/example"
	payload, err := Generate([]proxysite.ProxySite{site})
	if err != nil {
		t.Fatalf("生成静态站点失败: %v", err)
	}
	compact := compactJSON(payload)
	for _, expected := range []string{`"handler":"vars","root":"/var/www/example"`, `"handler":"file_server"`} {
		if !bytes.Contains(compact, []byte(expected)) {
			t.Fatalf("静态站点缺少 %s: %s", expected, payload)
		}
	}
	if bytes.Contains(compact, []byte(`"dial":"127.0.0.1:3000"`)) {
		t.Fatalf("静态站点不应包含业务上游: %s", payload)
	}
}

func TestGenerateSupportsSPAWithAPIAndCacheHeaders(t *testing.T) {
	site := testSite(true, []string{"127.0.0.1:3000"})
	site.SiteType = "spa"
	site.RootPath = "/var/www/app/dist"
	site.APIPath = "/api/*"
	site.EnableSecurityHeaders = true
	site.EnableAssetCache = true
	payload, err := Generate([]proxysite.ProxySite{site})
	if err != nil {
		t.Fatalf("生成 SPA 站点失败: %v", err)
	}
	compact := compactJSON(payload)
	for _, expected := range []string{
		`"path":["/api/*"]`, `"handler":"reverse_proxy"`, `"root":"/var/www/app/dist"`,
		`"try_files":["{http.request.uri.path}","/index.html"]`,
		`"X-Content-Type-Options":["nosniff"]`, `"Cache-Control":["public, max-age=31536000, immutable"]`,
	} {
		if !bytes.Contains(compact, []byte(expected)) {
			t.Fatalf("SPA 站点缺少 %s: %s", expected, payload)
		}
	}
}

func TestGenerateEnablesPerSiteAccessLog(t *testing.T) {
	t.Setenv("CADDYPILOT_LOG_DIR", "/data/logs")
	site := testSite(true, []string{"127.0.0.1:3000"})
	site.EnableLog = true
	payload, err := Generate([]proxysite.ProxySite{site})
	if err != nil {
		t.Fatalf("生成访问日志配置失败: %v", err)
	}
	compact := compactJSON(payload)
	for _, expected := range []string{`"http.log.access.sites"`, `"example.com":["sites"]`, `"filename":"/data/logs/access.log"`} {
		if !bytes.Contains(compact, []byte(expected)) {
			t.Fatalf("访问日志配置缺少 %s: %s", expected, payload)
		}
	}
}

func TestGenerateSupportsTypedUpstreams(t *testing.T) {
	tests := []struct {
		name         string
		upstreamType string
		upstream     string
		expected     []string
	}{
		{name: "HTTPS", upstreamType: "https", upstream: "10.0.0.8:443", expected: []string{`"tls":{}`, `"protocol":"http"`}},
		{name: "h2c", upstreamType: "h2c", upstream: "10.0.0.8:50051", expected: []string{`"versions":["h2c"]`}},
		{name: "Unix Socket", upstreamType: "unix", upstream: "/run/app.sock", expected: []string{`"dial":"unix//run/app.sock"`}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			site := testSite(true, []string{test.upstream})
			site.UpstreamType = test.upstreamType
			payload, err := Generate([]proxysite.ProxySite{site})
			if err != nil {
				t.Fatalf("生成类型化上游失败: %v", err)
			}
			compact := compactJSON(payload)
			for _, expected := range test.expected {
				if !bytes.Contains(compact, []byte(expected)) {
					t.Fatalf("生成配置缺少 %s: %s", expected, payload)
				}
			}
		})
	}
}

func TestGenerateSupportsUpstreamTLSOptions(t *testing.T) {
	site := testSite(true, []string{"10.0.0.8:443"})
	site.UpstreamType = "https"
	site.UpstreamTLSServerName = "backend.example.com"
	site.UpstreamTLSInsecureSkipVerify = true
	payload, err := Generate([]proxysite.ProxySite{site})
	if err != nil {
		t.Fatalf("生成 HTTPS 上游失败: %v", err)
	}
	compact := compactJSON(payload)
	for _, expected := range []string{`"server_name":"backend.example.com"`, `"insecure_skip_verify":true`} {
		if !bytes.Contains(compact, []byte(expected)) {
			t.Fatalf("生成配置缺少 %s: %s", expected, payload)
		}
	}
}

func TestGenerateAddsTLSConnectionPolicy(t *testing.T) {
	site := testSite(true, []string{"127.0.0.1:3000"})
	site.EnableHTTPS = true
	payload, err := Generate([]proxysite.ProxySite{site})
	if err != nil {
		t.Fatalf("生成 HTTPS 配置失败: %v", err)
	}
	if !bytes.Contains(payload, []byte("tls_connection_policies")) {
		t.Fatalf("HTTPS 服务缺少 TLS 连接策略: %s", payload)
	}
}

func TestGenerateRedirectUsesExternalHTTPSPort(t *testing.T) {
	t.Setenv("CADDYPILOT_HTTPS_PORT", "8443")
	site := testSite(true, []string{"127.0.0.1:3000"})
	site.EnableHTTPS = true
	site.ForceHTTPS = true
	payload, err := Generate([]proxysite.ProxySite{site})
	if err != nil {
		t.Fatalf("生成 HTTPS 跳转失败: %v", err)
	}
	if !bytes.Contains(payload, []byte(`https://{http.request.host}:8443{http.request.uri}`)) {
		t.Fatalf("非标准 HTTPS 端口未进入跳转地址: %s", payload)
	}
}

func TestGenerateRedirectOmitsStandardHTTPSPort(t *testing.T) {
	t.Setenv("CADDYPILOT_HTTPS_PORT", "443")
	site := testSite(true, []string{"127.0.0.1:3000"})
	site.EnableHTTPS = true
	site.ForceHTTPS = true
	payload, err := Generate([]proxysite.ProxySite{site})
	if err != nil {
		t.Fatalf("生成 HTTPS 跳转失败: %v", err)
	}
	if bytes.Contains(payload, []byte(`{http.request.host}:443`)) {
		t.Fatalf("标准 HTTPS 端口不应写入跳转地址: %s", payload)
	}
}

func TestGenerateRejectsInvalidExternalHTTPSPort(t *testing.T) {
	t.Setenv("CADDYPILOT_HTTPS_PORT", "invalid")
	site := testSite(true, []string{"127.0.0.1:3000"})
	site.EnableHTTPS = true
	site.ForceHTTPS = true
	if _, err := Generate([]proxysite.ProxySite{site}); err == nil {
		t.Fatal("应拒绝无效的全局 HTTPS 端口")
	}
}

func TestGenerateSupportsAliDNSWildcardCertificate(t *testing.T) {
	site := testSite(true, []string{"127.0.0.1:3000"})
	site.EnableHTTPS = true
	site.CertificateType = "wildcard"
	site.CertificateDomain = "*.example.com"
	site.ACMEChallengeType = "dns"
	site.DNSProvider = "alidns"
	payload, err := Generate([]proxysite.ProxySite{site})
	if err != nil {
		t.Fatalf("生成阿里云 DNS-01 配置失败: %v", err)
	}
	compact := compactJSON(payload)
	for _, expected := range []string{`"subjects":["*.example.com"]`, `"automate":["*.example.com"]`, `"name":"alidns"`, `{env.ALIYUN_ACCESS_KEY_ID}`, `{env.ALIYUN_ACCESS_KEY_SECRET}`} {
		if !bytes.Contains(compact, []byte(expected)) {
			t.Fatalf("DNS-01 配置缺少 %s: %s", expected, payload)
		}
	}
}

func TestGenerateMergesDuplicateWildcardTLSPolicies(t *testing.T) {
	first := testSite(true, []string{"127.0.0.1:3000"})
	first.Name = "站点一"
	first.EnableHTTPS = true
	first.CertificateType = "wildcard"
	first.ACMEChallengeType = "dns"
	first.ResolvedCertificateSubjects = mustJSON([]string{"*.example.com"})
	second := first
	second.Name = "站点二"
	second.Domains = mustJSON([]string{"app.example.com"})

	payload, err := Generate([]proxysite.ProxySite{first, second})
	if err != nil {
		t.Fatalf("合并重复 TLS 策略失败: %v", err)
	}
	if bytes.Count(payload, []byte(`"*.example.com"`)) != 2 {
		t.Fatalf("通配符 TLS 策略未正确生成: %s", payload)
	}
}

func TestGenerateRejectsConflictingWildcardTLSPolicies(t *testing.T) {
	first := testSite(true, []string{"127.0.0.1:3000"})
	first.Name = "站点一"
	first.EnableHTTPS = true
	first.CertificateType = "wildcard"
	first.ACMEChallengeType = "dns"
	first.ResolvedCertificateSubjects = mustJSON([]string{"*.example.com"})
	firstProvider := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	first.DNSProviderID = &firstProvider
	second := first
	second.Name = "站点二"
	second.Domains = mustJSON([]string{"app.example.com"})
	secondProvider := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	second.DNSProviderID = &secondProvider

	_, err := Generate([]proxysite.ProxySite{first, second})
	if err == nil || !bytes.Contains([]byte(err.Error()), []byte("不同的 TLS 自动化策略")) {
		t.Fatalf("应拒绝同域名的冲突 TLS 策略: %v", err)
	}
}

func TestGenerateUsesStoredDNSProviderEnvironment(t *testing.T) {
	site := testSite(true, []string{"127.0.0.1:3000"})
	providerID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	site.EnableHTTPS = true
	site.CertificateType = "single"
	site.ACMEChallengeType = "dns"
	site.DNSProviderID = &providerID
	payload, err := Generate([]proxysite.ProxySite{site})
	if err != nil {
		t.Fatalf("生成系统 DNS Provider 配置失败: %v", err)
	}
	idName, secretName, _ := dnsprovider.EnvNames(providerID)
	for _, expected := range []string{"{env." + idName + "}", "{env." + secretName + "}"} {
		if !bytes.Contains(payload, []byte(expected)) {
			t.Fatalf("配置缺少动态凭据占位符 %s: %s", expected, payload)
		}
	}
}

func TestEnsureManagementEntryInjectsProtectedServer(t *testing.T) {
	payload, err := EnsureManagementEntry([]byte(`{"apps":{"http":{"servers":{}}}}`))
	if err != nil {
		t.Fatalf("注入管理入口失败: %v", err)
	}
	if !HasManagementEntry(payload) {
		t.Fatalf("注入后仍缺少管理入口: %s", payload)
	}
}

func TestGenerateSupportsDevelopmentFrontendProxy(t *testing.T) {
	t.Setenv("CADDYPILOT_FRONTEND_PROXY", "http://127.0.0.1:3000")
	payload, err := Generate(nil)
	if err != nil {
		t.Fatalf("生成开发入口配置失败: %v", err)
	}
	if !bytes.Contains(payload, []byte("127.0.0.1:3000")) || !HasManagementEntry(payload) {
		t.Fatalf("开发入口未代理到 Vite: %s", payload)
	}
}

func testSite(enabled bool, upstreams []string) proxysite.ProxySite {
	return proxysite.ProxySite{
		Name:            "测试站点",
		Domains:         mustJSON([]string{"example.com"}),
		Upstreams:       mustJSON(upstreams),
		RequestHeaders:  mustJSON(map[string]string{}),
		ResponseHeaders: mustJSON(map[string]string{}),
		BasicAuthUsers:  mustJSON(map[string]string{}),
		AllowedIPs:      mustJSON([]string{}),
		EnableGzip:      true,
		Enabled:         enabled,
	}
}

func mustJSON(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(payload)
}

func compactJSON(payload []byte) []byte {
	var buffer bytes.Buffer
	if err := json.Compact(&buffer, payload); err != nil {
		panic(err)
	}
	return buffer.Bytes()
}
