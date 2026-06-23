package logs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadEntriesSupportsInitialTailAndCursor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "system.log")
	payload := "{\"ts\":\"2026-06-22T12:00:00Z\",\"level\":\"info\",\"msg\":\"one\"}\n" +
		"{\"ts\":\"2026-06-22T12:00:01Z\",\"level\":\"error\",\"msg\":\"two\"}\n"
	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	entries, cursor, err := readEntries(file, 0, 1)
	file.Close()
	if err != nil || len(entries) != 1 || entries[0].Message != "two" || cursor != int64(len(payload)) {
		t.Fatalf("初始日志尾部读取失败: %+v, %d, %v", entries, cursor, err)
	}

	file, _ = os.Open(path)
	entries, next, err := readEntries(file, cursor, 10)
	file.Close()
	if err != nil || len(entries) != 0 || next != cursor {
		t.Fatalf("游标增量读取失败: %+v, %d, %v", entries, next, err)
	}
}

func TestDNSLogFilteringAndCredentialSanitizing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "caddy.log")
	payload := "{\"ts\":1,\"level\":\"info\",\"msg\":\"unrelated\"}\n" +
		"{\"ts\":2,\"level\":\"info\",\"msg\":\"dns_provider_call\",\"zone\":\"example.com.\",\"records\":[{\"name\":\"_acme-challenge\",\"type\":\"TXT\",\"value\":\"token-value\"}],\"access_key_secret\":\"must-not-leak\",\"authorization\":\"must-not-leak\"}\n"
	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	entries, _, err := readFilteredEntries(file, 0, 10, func(entry Entry) bool {
		return entry.Message == "dns_provider_call"
	})
	file.Close()
	if err != nil || len(entries) != 1 {
		t.Fatalf("DNS 日志筛选失败: %+v, %v", entries, err)
	}
	if entries[0].Fields["zone"] != "example.com." || entries[0].Fields["access_key_secret"] != nil || entries[0].Fields["authorization"] != nil {
		t.Fatalf("DNS 日志字段或脱敏不正确: %+v", entries[0].Fields)
	}
}
