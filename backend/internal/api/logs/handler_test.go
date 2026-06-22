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
