package logs

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

func List(_ context.Context, input *ListInput) (*ListOutput, error) {
	filename := input.Source
	if input.Source == "dns" {
		filename = "caddy"
	}
	path := filepath.Join(logDirectory(), filename+".log")
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return &ListOutput{Body: ListResponse{Entries: []Entry{}}}, nil
	}
	if err != nil {
		return nil, huma.Error500InternalServerError("打开日志文件失败", err)
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, huma.Error500InternalServerError("读取日志状态失败", err)
	}
	cursor := input.Cursor
	if cursor < 0 || cursor > stat.Size() {
		cursor = 0
	}
	entries, nextCursor, err := readFilteredEntries(file, cursor, input.Limit, func(entry Entry) bool {
		return input.Source != "dns" || entry.Message == "dns_provider_call"
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("读取日志失败", err)
	}
	return &ListOutput{Body: ListResponse{Entries: entries, NextCursor: nextCursor}}, nil
}

func readEntries(file *os.File, cursor int64, limit int) ([]Entry, int64, error) {
	return readFilteredEntries(file, cursor, limit, nil)
}

func readFilteredEntries(file *os.File, cursor int64, limit int, include func(Entry) bool) ([]Entry, int64, error) {
	if cursor > 0 {
		if _, err := file.Seek(cursor, io.SeekStart); err != nil {
			return nil, cursor, err
		}
	}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1<<20)
	entries := make([]Entry, 0, limit)
	offset := cursor
	for scanner.Scan() {
		line := scanner.Text()
		offset += int64(len(scanner.Bytes()) + 1)
		entry := parseEntry(line, offset)
		if include != nil && !include(entry) {
			continue
		}
		if cursor == 0 && len(entries) == limit {
			copy(entries, entries[1:])
			entries[len(entries)-1] = entry
			continue
		}
		if len(entries) < limit {
			entries = append(entries, entry)
		}
	}
	for left, right := 0, len(entries)-1; left < right; left, right = left+1, right-1 {
		entries[left], entries[right] = entries[right], entries[left]
	}
	return entries, offset, scanner.Err()
}

func parseEntry(line string, offset int64) Entry {
	entry := Entry{ID: strconv.FormatInt(offset, 10), Message: strings.TrimSpace(line)}
	var payload map[string]any
	if json.Unmarshal([]byte(line), &payload) != nil {
		return entry
	}
	if level, ok := payload["level"].(string); ok {
		entry.Level = strings.ToUpper(level)
	}
	if message, ok := payload["msg"].(string); ok && message != "" {
		entry.Message = message
	}
	entry.Timestamp = formatTimestamp(payload["ts"])
	delete(payload, "ts")
	delete(payload, "level")
	delete(payload, "msg")
	entry.Fields = sanitizeLogFields(payload)
	return entry
}

func sanitizeLogFields(fields map[string]any) map[string]any {
	sanitized := make(map[string]any, len(fields))
	for key, value := range fields {
		if sensitiveLogKey(key) {
			continue
		}
		sanitized[key] = sanitizeLogValue(value)
	}
	return sanitized
}

func sanitizeLogValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return sanitizeLogFields(typed)
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, sanitizeLogValue(item))
		}
		return result
	default:
		return value
	}
}

func sensitiveLogKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "_", ""), "-", ""))
	for _, fragment := range []string{"accesskey", "secret", "securitytoken", "signature", "authorization", "authheader"} {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}

func formatTimestamp(value any) string {
	switch timestamp := value.(type) {
	case string:
		return timestamp
	case float64:
		seconds := int64(timestamp)
		nanoseconds := int64((timestamp - float64(seconds)) * float64(time.Second))
		return time.Unix(seconds, nanoseconds).Format(time.RFC3339)
	default:
		return ""
	}
}

func logDirectory() string {
	if value := strings.TrimSpace(os.Getenv("CADDYPILOT_LOG_DIR")); value != "" {
		return value
	}
	return filepath.Join("data", "logs")
}
