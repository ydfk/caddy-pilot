package alidns

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	libdnsalidns "github.com/libdns/alidns"
	"github.com/libdns/libdns"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestAuditLogContainsBusinessFieldsWithoutCredentials(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	provider := &Provider{Provider: new(libdnsalidns.Provider), logger: zap.New(core)}
	provider.AccessKeyID = "access-key-id"
	provider.AccessKeySecret = "access-key-secret"
	provider.SecurityToken = "security-token"
	records := []libdns.Record{libdns.TXT{Name: "_acme-challenge", Text: "challenge-value"}}

	provider.log("append", "example.com.", records, records, time.Now(), nil)

	entries := logs.AllUntimed()
	if len(entries) != 1 {
		t.Fatalf("期望 1 条审计日志，实际为 %d", len(entries))
	}
	payload, err := json.Marshal(entries[0].ContextMap())
	if err != nil {
		t.Fatalf("编码审计日志失败: %v", err)
	}
	content := string(payload)
	for _, expected := range []string{"example.com.", "append", "_acme-challenge", "TXT", "challenge-value"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("审计日志缺少 %q: %s", expected, content)
		}
	}
	for _, secret := range []string{"access-key-id", "access-key-secret", "security-token"} {
		if strings.Contains(content, secret) {
			t.Fatalf("审计日志泄露凭据 %q", secret)
		}
	}
}
