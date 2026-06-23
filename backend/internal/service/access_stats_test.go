package service

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadAccessStats(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("CADDYPILOT_LOG_DIR", directory)
	now := float64(time.Now().Unix())
	payload := fmt.Sprintf("{\"ts\":%f,\"status\":200,\"size\":120,\"request\":{\"host\":\"app.example.com:443\"}}\n", now) +
		fmt.Sprintf("{\"ts\":%f,\"status\":502,\"size\":30,\"request\":{\"host\":\"app.example.com\"}}\n", now) +
		fmt.Sprintf("{\"ts\":%f,\"status\":200,\"size\":10,\"request\":{\"host\":\"other.example.com\"}}\n", now-172800)
	if err := os.WriteFile(filepath.Join(directory, "access.log"), []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	stats, err := LoadAccessStats(time.Now().Add(-24*time.Hour), 5)
	if err != nil {
		t.Fatal(err)
	}
	if stats.RequestCount != 2 || stats.ErrorCount != 1 || stats.Bytes != 150 || len(stats.TopSites) != 1 {
		t.Fatalf("访问统计不正确: %+v", stats)
	}
	if stats.TopSites[0].Domain != "app.example.com" || stats.TopSites[0].RequestCount != 2 {
		t.Fatalf("热门站点不正确: %+v", stats.TopSites)
	}
}
