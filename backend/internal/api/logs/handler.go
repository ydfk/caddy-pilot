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
	path := filepath.Join(logDirectory(), input.Source+".log")
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
	entries, nextCursor, err := readEntries(file, cursor, input.Limit)
	if err != nil {
		return nil, huma.Error500InternalServerError("读取日志失败", err)
	}
	return &ListOutput{Body: ListResponse{Entries: entries, NextCursor: nextCursor}}, nil
}

func readEntries(file *os.File, cursor int64, limit int) ([]Entry, int64, error) {
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
		if cursor == 0 && len(entries) == limit {
			copy(entries, entries[1:])
			entries[len(entries)-1] = entry
			continue
		}
		if len(entries) < limit {
			entries = append(entries, entry)
		}
	}
	return entries, offset, scanner.Err()
}

func parseEntry(line string, offset int64) Entry {
	entry := Entry{ID: strconv.FormatInt(offset, 10), Message: strings.TrimSpace(line)}
	var payload struct {
		Timestamp any    `json:"ts"`
		Level     string `json:"level"`
		Message   string `json:"msg"`
	}
	if json.Unmarshal([]byte(line), &payload) != nil {
		return entry
	}
	entry.Level = strings.ToUpper(payload.Level)
	if payload.Message != "" {
		entry.Message = payload.Message
	}
	entry.Timestamp = formatTimestamp(payload.Timestamp)
	return entry
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
