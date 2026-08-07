package systeminfo

import "testing"

func TestInfoUsesConfiguredPublicPorts(t *testing.T) {
	t.Setenv("CADDYPILOT_HTTP_PORT", "18080")
	t.Setenv("CADDYPILOT_HTTPS_PORT", "18443")
	output, err := Info(t.Context(), &struct{}{})
	if err != nil {
		t.Fatalf("读取系统信息失败: %v", err)
	}
	if output.Body.HTTPPort != 18080 || output.Body.HTTPSPort != 18443 {
		t.Fatalf("公开端口错误: %+v", output.Body)
	}
}

func TestInfoRejectsInvalidPublicPort(t *testing.T) {
	t.Setenv("CADDYPILOT_HTTP_PORT", "invalid")
	if _, err := Info(t.Context(), &struct{}{}); err == nil {
		t.Fatal("无效公开端口应返回错误")
	}
}
