package caddygen

import (
	"encoding/json"
	"fmt"

	"go-fiber-starter/internal/model/proxysite"
)

func GenerateSiteRoute(site proxysite.ProxySite) (map[string]any, error) {
	domains, err := decodeStringList(site.Domains)
	if err != nil {
		return nil, fmt.Errorf("解析域名失败: %w", err)
	}
	upstreamValues, err := decodeStringList(site.Upstreams)
	if err != nil {
		return nil, fmt.Errorf("解析上游失败: %w", err)
	}

	upstreams := make([]map[string]string, 0, len(upstreamValues))
	for _, upstream := range upstreamValues {
		upstreams = append(upstreams, map[string]string{"dial": upstream})
	}

	return map[string]any{
		"match": []map[string]any{{"host": domains}},
		"handle": []map[string]any{{
			"handler":   "reverse_proxy",
			"upstreams": upstreams,
		}},
	}, nil
}

func decodeStringList(value string) ([]string, error) {
	var items []string
	if err := json.Unmarshal([]byte(value), &items); err != nil {
		return nil, err
	}
	return items, nil
}
