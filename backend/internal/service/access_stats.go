package service

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type SiteAccessStats struct {
	Domain       string
	RequestCount int64
	ErrorCount   int64
	Bytes        int64
}

type AccessStats struct {
	RequestCount int64
	ErrorCount   int64
	Bytes        int64
	TopSites     []SiteAccessStats
}

type accessLogEntry struct {
	Timestamp float64 `json:"ts"`
	Status    int     `json:"status"`
	Size      int64   `json:"size"`
	Request   struct {
		Host string `json:"host"`
	} `json:"request"`
}

func LoadAccessStats(since time.Time, topLimit int) (AccessStats, error) {
	path := filepath.Join(environmentValue("CADDYPILOT_LOG_DIR", filepath.Join("data", "logs")), "access.log")
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return AccessStats{TopSites: []SiteAccessStats{}}, nil
	}
	if err != nil {
		return AccessStats{}, err
	}
	defer file.Close()

	stats := AccessStats{TopSites: []SiteAccessStats{}}
	bySite := make(map[string]*SiteAccessStats)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1<<20)
	for scanner.Scan() {
		var entry accessLogEntry
		if json.Unmarshal(scanner.Bytes(), &entry) != nil || entry.Timestamp <= 0 || time.Unix(int64(entry.Timestamp), 0).Before(since) {
			continue
		}
		domain := normalizeAccessHost(entry.Request.Host)
		if domain == "" {
			continue
		}
		stats.RequestCount++
		stats.Bytes += entry.Size
		if entry.Status >= 500 {
			stats.ErrorCount++
		}
		site := bySite[domain]
		if site == nil {
			site = &SiteAccessStats{Domain: domain}
			bySite[domain] = site
		}
		site.RequestCount++
		site.Bytes += entry.Size
		if entry.Status >= 500 {
			site.ErrorCount++
		}
	}
	if err := scanner.Err(); err != nil {
		return AccessStats{}, err
	}
	for _, site := range bySite {
		stats.TopSites = append(stats.TopSites, *site)
	}
	sort.Slice(stats.TopSites, func(left, right int) bool {
		if stats.TopSites[left].RequestCount == stats.TopSites[right].RequestCount {
			return stats.TopSites[left].Domain < stats.TopSites[right].Domain
		}
		return stats.TopSites[left].RequestCount > stats.TopSites[right].RequestCount
	})
	if topLimit > 0 && len(stats.TopSites) > topLimit {
		stats.TopSites = stats.TopSites[:topLimit]
	}
	return stats, nil
}

func normalizeAccessHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if index := strings.LastIndex(host, ":"); index > 0 && !strings.Contains(host[index+1:], "]") {
		host = host[:index]
	}
	return strings.Trim(host, "[]")
}
