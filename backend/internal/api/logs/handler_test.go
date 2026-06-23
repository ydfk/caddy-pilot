package logs

import (
	"context"
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

func TestListFiltersDNSLogsByProvider(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("CADDYPILOT_LOG_DIR", directory)
	payload := "{\"msg\":\"dns_provider_call\",\"provider_id\":\"provider-1\",\"zone\":\"one.example.\"}\n" +
		"{\"msg\":\"dns_provider_call\",\"provider_id\":\"provider-2\",\"zone\":\"two.example.\"}\n"
	if err := os.WriteFile(filepath.Join(directory, "caddy.log"), []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	output, err := List(context.Background(), &ListInput{Source: "dns", ProviderID: "provider-2", Limit: 10})
	if err != nil || len(output.Body.Entries) != 1 || output.Body.Entries[0].Fields["zone"] != "two.example." {
		t.Fatalf("DNS Provider 日志筛选失败: %+v, %v", output, err)
	}
}

func TestReadEntriesReturnsNewestFirst(t *testing.T) {
	path := filepath.Join(t.TempDir(), "system.log")
	payload := "{\"msg\":\"old\"}\n{\"msg\":\"new\"}\n"
	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	file, _ := os.Open(path)
	entries, _, err := readEntries(file, 0, 10)
	file.Close()
	if err != nil || len(entries) != 2 || entries[0].Message != "new" || entries[1].Message != "old" {
		t.Fatalf("日志未按新到旧返回: %+v, %v", entries, err)
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
