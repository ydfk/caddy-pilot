package service

import "testing"

func TestCaddyErrorMessageKeepsFinalError(t *testing.T) {
	output := []byte("{\"level\":\"info\",\"msg\":\"loading\"}\nError: loading module: unknown module")
	if got := caddyErrorMessage(output); got != "loading module: unknown module" {
		t.Fatalf("Caddy 错误提取不正确: %q", got)
	}
}
