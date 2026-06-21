package caddygen

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

const (
	ManagementServerName = "caddypilot-admin"
	ManagementListen     = ":8080"
	BackendUpstream      = "127.0.0.1:25610"
	FrontendRoot         = "/app/frontend"
)

func EnsureManagementEntry(payload []byte) ([]byte, error) {
	var config map[string]any
	if err := json.Unmarshal(payload, &config); err != nil {
		return nil, err
	}
	servers, err := serverMap(config)
	if err != nil {
		return nil, err
	}
	for name, value := range servers {
		server, ok := value.(map[string]any)
		if ok && serverListensOn(server, managementListen()) {
			delete(servers, name)
		}
	}
	servers[ManagementServerName] = managementServer()
	config["admin"] = localAdminConfig()
	return json.MarshalIndent(config, "", "  ")
}

func HasManagementEntry(payload []byte) bool {
	var config map[string]any
	if json.Unmarshal(payload, &config) != nil {
		return false
	}
	servers, err := serverMap(config)
	if err != nil {
		return false
	}
	for _, value := range servers {
		server, ok := value.(map[string]any)
		if !ok || !serverListensOn(server, managementListen()) {
			continue
		}
		encoded, err := json.Marshal(server)
		if err == nil && containsAll(encoded, managementMarkers()) {
			return true
		}
	}
	return false
}

func managementServer() map[string]any {
	listen := managementListen()
	backend := backendUpstream()
	frontendProxy := normalizeDial(os.Getenv("CADDYPILOT_FRONTEND_PROXY"))
	routes := []map[string]any{
		{
			"group": "caddypilot-management",
			"match": []map[string]any{{"path": []string{"/api/*"}}},
			"handle": []map[string]any{{
				"handler": "subroute",
				"routes": []map[string]any{{"handle": []map[string]any{
					reverseProxyHandler([]string{backend}, nil),
				}}},
			}},
		},
	}
	if frontendProxy != "" {
		routes = append(routes, map[string]any{
			"group":  "caddypilot-management",
			"handle": []map[string]any{reverseProxyHandler([]string{frontendProxy}, nil)},
		})
	} else {
		routes = append([]map[string]any{{
			"handle": []map[string]any{
				{"handler": "vars", "root": frontendRoot()},
				encodeHandler(),
			},
		}}, routes...)
		routes = append(routes, map[string]any{
			"group": "caddypilot-management",
			"handle": []map[string]any{{
				"handler": "subroute",
				"routes": []map[string]any{
					{
						"match": []map[string]any{{"file": map[string]any{
							"try_files": []string{"{http.request.uri.path}", "/index.html"},
						}}},
						"handle": []map[string]any{{"handler": "rewrite", "uri": "{http.matchers.file.relative}"}},
					},
					{"handle": []map[string]any{{"handler": "file_server"}}},
				},
			}},
		})
	}
	return map[string]any{
		"listen":          []string{listen},
		"routes":          routes,
		"automatic_https": map[string]any{"disable_redirects": true},
	}
}

func managementMarkers() []string {
	markers := []string{"/api/*", backendUpstream()}
	if proxy := normalizeDial(os.Getenv("CADDYPILOT_FRONTEND_PROXY")); proxy != "" {
		return append(markers, proxy)
	}
	return append(markers, frontendRoot(), "file_server")
}

func managementListen() string {
	return environmentValue("CADDYPILOT_MANAGE_ADDR", ManagementListen)
}

func backendUpstream() string {
	return normalizeDial(environmentValue("CADDYPILOT_BACKEND_ADDR", BackendUpstream))
}

func frontendRoot() string {
	return environmentValue("CADDYPILOT_FRONTEND_DIR", FrontendRoot)
}

func normalizeDial(value string) string {
	return strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(value), "http://"), "https://")
}

func environmentValue(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func serverMap(config map[string]any) (map[string]any, error) {
	apps, ok := config["apps"].(map[string]any)
	if !ok {
		apps = map[string]any{}
		config["apps"] = apps
	}
	httpApp, ok := apps["http"].(map[string]any)
	if !ok {
		httpApp = map[string]any{}
		apps["http"] = httpApp
	}
	servers, ok := httpApp["servers"].(map[string]any)
	if !ok {
		servers = map[string]any{}
		httpApp["servers"] = servers
	}
	if servers == nil {
		return nil, errors.New("Caddy HTTP servers 无效")
	}
	return servers, nil
}

func serverListensOn(server map[string]any, address string) bool {
	listen, ok := server["listen"].([]any)
	if !ok {
		return false
	}
	for _, value := range listen {
		if value == address {
			return true
		}
	}
	return false
}

func containsAll(payload []byte, values []string) bool {
	text := string(payload)
	for _, value := range values {
		if !strings.Contains(text, value) {
			return false
		}
	}
	return true
}
