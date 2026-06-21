package caddygen

import (
	"bytes"
	"encoding/json"
	"testing"

	"go-fiber-starter/internal/model/proxysite"
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
